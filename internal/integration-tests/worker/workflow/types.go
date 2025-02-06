package integrationtest

import mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"

type DefaultTransformers struct {
	Boolean *mgmtv1alpha1.JobMappingTransformer
	String  *mgmtv1alpha1.JobMappingTransformer
	Number  *mgmtv1alpha1.JobMappingTransformer
	Byte    *mgmtv1alpha1.JobMappingTransformer
}

type TestJobOptions struct {
	SubsetByForeignKeyConstraints bool
	InitSchema                    bool
	Truncate                      bool
	TruncateCascade               bool
	DefaultTransformers           *DefaultTransformers
	SkipForeignKeyViolations      bool
	OnConflictDoNothing           bool
	OnConflictDoUpdate            bool
}
