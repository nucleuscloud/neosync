package v1alpha1_transformersservice

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
)

var (
	enLanguage = "en"
)

func (s *Service) GetTransformPiiRecognizers(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetTransformPiiRecognizersRequest],
) (*connect.Response[mgmtv1alpha1.GetTransformPiiRecognizersResponse], error) {
	if !s.cfg.IsPresidioEnabled {
		return nil, nucleuserrors.NewNotImplemented(fmt.Sprintf("%s is not implemented", strings.TrimPrefix(mgmtv1alpha1connect.TransformersServiceGetTransformPiiRecognizersProcedure, "/")))
	}
	if s.recognizerclient == nil {
		return nil, nucleuserrors.NewInternalError("recognizer service is enabled but client was nil.")
	}
	resp, err := s.recognizerclient.GetRecognizers(ctx, &presidioapi.GetRecognizersParams{
		Language: &enLanguage,
	})
	if err != nil {
		return nil, fmt.Errorf("was unable to retrieve available recognizers: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("received non-200 response from recognizer api: %s %d %s", resp.Status(), resp.StatusCode(), string(resp.Body))
	}

	recognizers := *resp.JSON200
	return connect.NewResponse(&mgmtv1alpha1.GetTransformPiiRecognizersResponse{
		Recognizers: recognizers,
	}), nil
}
