package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ToSha256(t *testing.T) {
	require.Equal(
		t,
		ToSha256("foobar"),
		"c3ab8ff13720e8ad9047dd39466b3c8974e592c2fa383d4a3960714caef0c4f2",
	)
}
