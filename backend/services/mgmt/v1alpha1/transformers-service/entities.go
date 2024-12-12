package v1alpha1_transformersservice

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/backend/internal/ee/rbac"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/userdata"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
)

var (
	enLanguage = "en"
)

func (s *Service) GetTransformPiiEntities(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetTransformPiiEntitiesRequest],
) (*connect.Response[mgmtv1alpha1.GetTransformPiiEntitiesResponse], error) {
	if !s.cfg.IsPresidioEnabled {
		return nil, nucleuserrors.NewNotImplemented(fmt.Sprintf("%s is not implemented", strings.TrimPrefix(mgmtv1alpha1connect.TransformersServiceGetTransformPiiEntitiesProcedure, "/")))
	}
	if s.entityclient == nil {
		return nil, nucleuserrors.NewInternalError("entity service is enabled but client was nil.")
	}
	user, err := s.userdataclient.GetUser(ctx)
	if err != nil {
		return nil, err
	}
	err = user.EnforceJob(ctx, userdata.NewWildcardDomainEntity(req.Msg.GetAccountId()), rbac.JobAction_View)
	if err != nil {
		return nil, err
	}

	resp, err := s.entityclient.GetSupportedentitiesWithResponse(ctx, &presidioapi.GetSupportedentitiesParams{
		Language: &enLanguage,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve available entities: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("received non-200 response from entity api: %s %d %s", resp.Status(), resp.StatusCode(), string(resp.Body))
	}

	entities := *resp.JSON200
	return connect.NewResponse(&mgmtv1alpha1.GetTransformPiiEntitiesResponse{
		Entities: entities,
	}), nil
}
