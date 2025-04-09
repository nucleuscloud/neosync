package jobmapping_builder_shared

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

func JobMappingsFromLegacyMappings(mappings []*mgmtv1alpha1.JobMapping) []*shared.JobTransformationMapping {
	jobMappings := make([]*shared.JobTransformationMapping, len(mappings))
	for i, mapping := range mappings {
		jobMappings[i] = jobMappingFromLegacyMapping(mapping)
	}
	return jobMappings
}

func jobMappingFromLegacyMapping(mapping *mgmtv1alpha1.JobMapping) *shared.JobTransformationMapping {
	return &shared.JobTransformationMapping{
		JobMapping:        mapping,
		DestinationSchema: mapping.GetSchema(),
		DestinationTable:  mapping.GetTable(),
	}
}
