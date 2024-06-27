package neosync_gcp

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"cloud.google.com/go/storage"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"google.golang.org/api/iterator"
)

type ClientInterface interface {
	GetDbSchemaFromPrefix(ctx context.Context, bucketName string, prefix string) ([]*mgmtv1alpha1.DatabaseColumn, error)
	DoesPrefixContainTables(ctx context.Context, bucketName string, prefix string) (bool, error)
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
		Prefix:    prefix,
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
		Prefix:    prefix,
		Delimiter: "/",
	})

	output := []*mgmtv1alpha1.DatabaseColumn{}
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

		columns, err := c.getTableColumnsFromFile(ctx, bucket, objAttrs.Prefix)
		if err != nil {
			return nil, err
		}
		for _, column := range columns {
			output = append(output, &mgmtv1alpha1.DatabaseColumn{
				Schema: schematable.Schema,
				Table:  schematable.Table,
				Column: column,
			})
		}
	}
	return output, nil
}

func (c *Client) getTableColumnsFromFile(
	ctx context.Context,
	bucket *storage.BucketHandle,
	prefix string,
) ([]string, error) {
	dataiterator := bucket.Objects(ctx, &storage.Query{Prefix: fmt.Sprintf("%s/data", strings.TrimSuffix(prefix, "/"))})
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
			c.logger.Warn(fmt.Sprintf("unable to successfully close gcs reader: %s", closeErr.Error()))
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
	return &sqlmanager_shared.SchemaTable{Schema: schemaTableList[0], Table: schemaTableList[1]}, nil
}

// Returns the prefix that contains the table folders in GCS
func GetWorkflowActivityPrefix(runId string, prefixPath *string) string {
	var pp = ""
	if prefixPath != nil {
		pp = strings.TrimSuffix(*prefixPath, "/")
	}
	return fmt.Sprintf("%s/workflows/%s/activities/", pp, runId)
}

func getFirstRecordFromReader(reader io.Reader) (map[string]any, error) {
	var result map[string]any

	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	scanner := bufio.NewScanner(gzipReader)
	if scanner.Scan() {
		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading from gzip reader: %w", err)
	}

	return result, nil
}
