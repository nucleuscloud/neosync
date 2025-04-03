package neosync_gcp

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"

	"cloud.google.com/go/storage"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/iterator"
)

type ClientInterface interface {
	GetDbSchemaFromPrefix(
		ctx context.Context,
		bucketName string,
		prefix string,
	) ([]*mgmtv1alpha1.DatabaseColumn, error)
	DoesPrefixContainTables(ctx context.Context, bucketName string, prefix string) (bool, error)
	GetRecordStreamFromPrefix(
		ctx context.Context,
		bucketName string,
		prefix string,
		onRecord func(record map[string][]byte) error,
	) error
	ListObjectPrefixes(ctx context.Context, bucketName, prefix, delimiter string) ([]string, error)
}

type Client struct {
	client *storage.Client
	logger *slog.Logger
}

var _ ClientInterface = &Client{}

func NewClient(client *storage.Client, logger *slog.Logger) *Client {
	return &Client{client: client, logger: logger}
}

func (c *Client) DoesPrefixContainTables(
	ctx context.Context,
	bucketName string,
	prefix string,
) (bool, error) {
	bucket := c.client.Bucket(bucketName)
	it := bucket.Objects(ctx, &storage.Query{
		Prefix:    fmt.Sprintf("%s/", strings.TrimSuffix(prefix, "/")),
		Delimiter: "/",
	})
	_, err := it.Next()
	if err != nil {
		if err == iterator.Done {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *Client) GetDbSchemaFromPrefix(
	ctx context.Context,
	bucketName string,
	prefix string,
) ([]*mgmtv1alpha1.DatabaseColumn, error) {
	bucket := c.client.Bucket(bucketName)
	it := bucket.Objects(ctx, &storage.Query{
		Prefix:    fmt.Sprintf("%s/", strings.TrimSuffix(prefix, "/")),
		Delimiter: "/",
	})

	output := []*mgmtv1alpha1.DatabaseColumn{}
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.SetLimit(10)
	mu := sync.Mutex{}

	for {
		objAttrs, err := it.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}
		if objAttrs.Prefix == "" {
			continue
		}
		schematable, err := getSchemaTableFromPrefix(objAttrs.Prefix)
		if err != nil {
			return nil, err
		}

		errgrp.Go(func() error {
			columns, err := c.getTableColumnsFromFile(errctx, bucket, objAttrs.Prefix)
			if err != nil {
				return err
			}
			mu.Lock()
			for _, column := range columns {
				output = append(output, &mgmtv1alpha1.DatabaseColumn{
					Schema: schematable.Schema,
					Table:  schematable.Table,
					Column: column,
				})
			}
			mu.Unlock()
			return nil
		})
	}
	err := errgrp.Wait()
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (c *Client) GetRecordStreamFromPrefix(
	ctx context.Context,
	bucketName string,
	prefix string,
	onRecord func(record map[string][]byte) error,
) error {
	bucket := c.client.Bucket(bucketName)
	it := bucket.Objects(ctx, &storage.Query{Prefix: prefix})

	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.SetLimit(10)
	for {
		objAttrs, err := it.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return err
		}
		if objAttrs.Name == "" {
			continue // encountered folder
		}
		errgrp.Go(func() error {
			reader, err := bucket.Object(objAttrs.Name).NewReader(errctx)
			if err != nil {
				return err
			}
			err = streamRecordsFromReader(reader, onRecord)
			if closeErr := reader.Close(); closeErr != nil {
				c.logger.Warn(
					fmt.Sprintf(
						"failed to close reader while streaming records from prefix: %s",
						closeErr.Error(),
					),
				)
			}
			return err
		})
	}
	return errgrp.Wait()
}

func (c *Client) ListObjectPrefixes(
	ctx context.Context,
	bucketName, prefix, delimiter string,
) ([]string, error) {
	prefixes := []string{}
	it := c.client.Bucket(bucketName).Objects(ctx, &storage.Query{
		Prefix:    prefix,
		Delimiter: delimiter,
	})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list objects from bucket %q: %w", bucketName, err)
		}
		prefixes = append(prefixes, attrs.Prefix)
	}
	return prefixes, nil
}

func (c *Client) getTableColumnsFromFile(
	ctx context.Context,
	bucket *storage.BucketHandle,
	prefix string,
) ([]string, error) {
	dataiterator := bucket.Objects(
		ctx,
		&storage.Query{Prefix: fmt.Sprintf("%s/data", strings.TrimSuffix(prefix, "/"))},
	)
	columns := []string{}

	var firstFile *storage.ObjectAttrs
	for {
		dataAttrs, err := dataiterator.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}
		if dataAttrs.Name == "" {
			continue
		}
		firstFile = dataAttrs
		break // we only care about the first record
	}
	if firstFile == nil {
		return nil, errors.New("unable to find a valid file for the given prefix")
	}
	reader, err := bucket.Object(firstFile.Name).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			c.logger.Warn(
				fmt.Sprintf("unable to successfully close gcs reader: %s", closeErr.Error()),
			)
		}
	}()

	firstRecord, err := getFirstRecordFromReader(reader)
	if err != nil {
		return nil, err
	}
	for column := range firstRecord {
		columns = append(columns, column)
	}
	return columns, nil
}

func getSchemaTableFromPrefix(prefix string) (*sqlmanager_shared.SchemaTable, error) {
	folders := strings.Split(prefix, "activities")
	tableFolder := strings.ReplaceAll(folders[len(folders)-1], "/", "")
	schemaTableList := strings.Split(tableFolder, ".")
	if len(schemaTableList) == 0 {
		return nil, errors.New("unable to parse schema table from prefix")
	}
	if len(schemaTableList) == 1 {
		return &sqlmanager_shared.SchemaTable{Schema: "", Table: schemaTableList[0]}, nil
	}
	return &sqlmanager_shared.SchemaTable{
		Schema: schemaTableList[0],
		Table:  schemaTableList[1],
	}, nil
}

// Returns the prefix that contains the table folders in GCS
func GetWorkflowActivityPrefix(runId string, prefixPath *string) string {
	var pp = ""
	if prefixPath != nil && *prefixPath != "" {
		pp = fmt.Sprintf("%s/", strings.TrimSuffix(*prefixPath, "/"))
	}
	return fmt.Sprintf("%sworkflows/%s/activities", pp, runId)
}

func GetWorkflowActivityDataPrefix(runId, table string, prefixPath *string) string {
	return fmt.Sprintf("%s/%s/data", GetWorkflowActivityPrefix(runId, prefixPath), table)
}

func getFirstRecordFromReader(reader io.Reader) (map[string]any, error) {
	var result map[string]any

	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	decoder := json.NewDecoder(gzipReader)
	err = decoder.Decode(&result)
	if err != nil && err == io.EOF {
		return result, nil
	} else if err != nil {
		return nil, err
	}

	return result, nil
}

func streamRecordsFromReader(
	reader io.Reader,
	onRecord func(record map[string][]byte) error,
) error {
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	decoder := json.NewDecoder(gzipReader)
	for {
		var result map[string]any
		err = decoder.Decode(&result)
		if err != nil && err == io.EOF {
			break // End of file, stop the loop
		} else if err != nil {
			return err
		}

		record, err := valToRecord(result)
		if err != nil {
			return fmt.Errorf(
				"unable to convert record from map[string]any to map[string][]byte: %w",
				err,
			)
		}
		err = onRecord(record)
		if err != nil {
			return err
		}
	}

	return nil
}

func valToRecord(input map[string]any) (map[string][]byte, error) {
	output := make(map[string][]byte)
	for key, value := range input {
		var byteValue []byte
		if str, ok := value.(string); ok {
			// try converting string directly to []byte
			// prevents quoted strings
			byteValue = []byte(str)
		} else {
			// if not a string use JSON encoding
			bits, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			if string(bits) == "null" {
				byteValue = nil
			} else {
				byteValue = bits
			}
		}
		output[key] = byteValue
	}
	return output, nil
}
