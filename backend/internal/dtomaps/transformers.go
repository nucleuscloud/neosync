package dtomaps

import (
	"fmt"

	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/internal/neosyncdb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToUserDefinedTransformerDto(
	input *db_queries.NeosyncApiTransformer,
	systemTransformers map[mgmtv1alpha1.TransformerSource]*mgmtv1alpha1.SystemTransformer,
) (*mgmtv1alpha1.UserDefinedTransformer, error) {
	if _, ok := mgmtv1alpha1.TransformerSource_name[input.Source]; !ok {
		return nil, fmt.Errorf("%d is not a valid transformer source", input.Source)
	}
	source := mgmtv1alpha1.TransformerSource(input.Source)
	transformer, ok := systemTransformers[source]
	if !ok {
		return nil, fmt.Errorf("source %d is valid, but was not found in system transformers map", input.Source)
	}
	return &mgmtv1alpha1.UserDefinedTransformer{
		Id:          neosyncdb.UUIDString(input.ID),
		Name:        input.Name,
		Description: input.Description,
		Source:      source,
		DataType:    transformer.DataType, //nolint:staticcheck
		DataTypes:   transformer.DataTypes,
		Config:      input.TransformerConfig.ToTransformerConfigDto(),
		CreatedAt:   timestamppb.New(input.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(input.UpdatedAt.Time),
		AccountId:   neosyncdb.UUIDString(input.AccountID),
	}, nil
}
