package transformers

import (
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

func Test_TransformPiiText(t *testing.T) {
	env := bloblang.NewEmptyEnvironment()
	mockanalyze := presidioapi.NewMockAnalyzeInterface(t)
	mockanon := presidioapi.NewMockAnonymizeInterface(t)
	err := NewTransformPiiText(env, mockanalyze, mockanon, &mgmtv1alpha1.TransformPiiText{})
	require.NoError(t, err)

	mockanalyze.On("PostAnalyzeWithResponse", mock.Anything, mock.Anything).
		Return(&presidioapi.PostAnalyzeResponse{
			JSON200: &[]presidioapi.RecognizerResultWithAnaysisExplanation{
				{
					Start:      13,
					End:        21,
					EntityType: "name",
					Score:      100,
				},
			},
		}, nil)

	mockText := "my name is asdf and I am 100 years old"
	mockanon.On("PostAnonymizeWithResponse", mock.Anything, mock.Anything).
		Return(&presidioapi.PostAnonymizeResponse{
			JSON200: &presidioapi.AnonymizeResponse{Text: &mockText},
		}, nil)

	exec, err := env.Parse(`root = transform_pii_text(value:"my name is john doe and I am 100 years old")`)
	require.NoError(t, err)
	output, err := exec.Query(nil)
	require.NoError(t, err)

	value, ok := output.(string)
	require.True(t, ok)
	require.Equal(t, mockText, value)
}
