package benthosbuilder_builders

import (
	"context"
	"errors"
	"strings"

	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

type gcpCloudStorageSyncBuilder struct {
}

func NewGcpCloudStorageSyncBuilder() bb_internal.BenthosBuilder {
	return &gcpCloudStorageSyncBuilder{}
}

func (b *gcpCloudStorageSyncBuilder) BuildSourceConfigs(ctx context.Context, params *bb_internal.SourceParams) ([]*bb_internal.BenthosSourceConfig, error) {
	return nil, errors.ErrUnsupported
}

func (b *gcpCloudStorageSyncBuilder) BuildDestinationConfig(ctx context.Context, params *bb_internal.DestinationParams) (*bb_internal.BenthosDestinationConfig, error) {
	config := &bb_internal.BenthosDestinationConfig{}

	benthosConfig := params.SourceConfig
	if benthosConfig.RunType == tabledependency.RunTypeUpdate {
		return config, nil
	}
	destinationOpts := params.DestinationOpts.GetAwsS3Options()
	gcpCloudStorageConfig := params.DestConnection.GetConnectionConfig().GetGcpCloudstorageConfig()

	if destinationOpts == nil {
		return nil, errors.New("destination must have configured GCP Cloud Storage options")
	}
	if gcpCloudStorageConfig == nil {
		return nil, errors.New("destination must have configured GCP Cloud Storage config")
	}

	pathpieces := []string{}
	if gcpCloudStorageConfig.GetPathPrefix() != "" {
		pathpieces = append(pathpieces, strings.Trim(gcpCloudStorageConfig.GetPathPrefix(), "/"))
	}

	pathpieces = append(
		pathpieces,
		"workflows",
		params.RunId,
		"activities",
		neosync_benthos.BuildBenthosTable(benthosConfig.TableSchema, benthosConfig.TableName),
		"data",
		`${!count("files")}.txt.gz`,
	)

	config.Outputs = append(config.Outputs, neosync_benthos.Outputs{
		Fallback: []neosync_benthos.Outputs{
			{
				GcpCloudStorage: &neosync_benthos.GcpCloudStorageOutput{
					Bucket:          gcpCloudStorageConfig.GetBucket(),
					MaxInFlight:     64,
					Path:            strings.Join(pathpieces, "/"),
					ContentType:     shared.Ptr("txt/plain"),
					ContentEncoding: shared.Ptr("gzip"),
					Batching: &neosync_benthos.Batching{
						Count:  100,
						Period: "5s",
						Processors: []*neosync_benthos.BatchProcessor{
							{Archive: &neosync_benthos.ArchiveProcessor{Format: "lines"}},
							{Compress: &neosync_benthos.CompressProcessor{Algorithm: "gzip"}},
						},
					},
				},
			},
			// kills activity depending on error
			{Error: &neosync_benthos.ErrorOutputConfig{
				ErrorMsg: `${! meta("fallback_error")}`,
				Batching: &neosync_benthos.Batching{
					Period: "5s",
					Count:  100,
				},
			}},
		},
	})

	return config, nil
}
