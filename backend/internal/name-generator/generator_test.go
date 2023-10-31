package namegenerator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NameGenerator_New(t *testing.T) {
	assert.NotNil(t, New(2, "-"))
}

func Test_NameGenerator_Generate(t *testing.T) {
	client := New(2, "-")

	output := client.Generate()
	assert.NotEmpty(t, output)
	splits := strings.Split(output, "-")
	assert.Len(t, splits, 2)
}

func Test_NameGenerator_Generate_DifferentNames(t *testing.T) {
	client := New(2, "-")

	o1 := client.Generate()
	o2 := client.Generate()
	assert.True(t, o1 != o2)
}
