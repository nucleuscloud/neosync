package transformers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/assert"
)

var fullName = "john smith"

func Test_TranformFullNameEmptyName(t *testing.T) {
	randomizer := rng.New(1)
	emptyName := ""

	res, err := transformFullName(randomizer, emptyName, true, maxCharacterLimit)
	assert.NoError(t, err)
	assert.Nil(t, res, "The response should be nil")
}

func Test_TransformFullNamePreserveLengthTrue(t *testing.T) {
	randomizer := rng.New(1)

	nameLength := int64(len(fullName))

	res, err := transformFullName(randomizer, fullName, true, maxCharacterLimit)

	assert.NoError(t, err)
	assert.Equal(t, nameLength, int64(len(*res)), "The first name output should be the same length as the input")
	assert.IsType(t, "", *res, "The first name should be a string")
}

func Test_TransformFullNameMaxLengthBetween12And5(t *testing.T) {
	randomizer := rng.New(1)
	res, err := transformFullName(randomizer, fullName, false, 10)

	assert.NoError(t, err)
	assert.True(t, len(*res) >= 6, "The name should be greater than the min length name")
	assert.True(t, len(*res) <= 10, "The name should be less than the max character limit")
	assert.IsType(t, "", *res, "The first name should be a string")
}

func Test_FullNameTransformerWithValue(t *testing.T) {
	fn := "john smith"
	mapping := fmt.Sprintf(`root = transform_full_name(value:%q,preserve_length:true,max_length:%d)`, fn, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the full name transformer")

	res, err := ex.Query(fn)
	assert.NoError(t, err)

	assert.NotNil(t, res, "The response shouldn't be nil.")

	resStr, ok := res.(*string)
	if !ok {
		t.Errorf("Expected *string, got %T", res)
		return
	}

	if resStr != nil {
		assert.Equal(t, len(*resStr), len(fn), "Generated full name must be as long as input full name")
	} else {
		t.Error("Pointer is nil, expected a valid string pointer")
	}
}

func Test_TransformFullNamelTransformerWithEmptyValue(t *testing.T) {
	nilName := ""
	mapping := fmt.Sprintf(`root = transform_full_name(value:%q,preserve_length:true,max_length:%d)`, nilName, maxCharacterLimit)
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the full name transformer")

	_, err = ex.Query(nil)
	assert.NoError(t, err)
}

func Test_TransformFullNamelTransformer_NoOptions(t *testing.T) {
	mapping := fmt.Sprintf(`root = transform_full_name(value:%q)`, "full name")
	ex, err := bloblang.Parse(mapping)
	assert.NoError(t, err, "failed to parse the full name transformer")

	res, err := ex.Query(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}

func Test_TransformFullNamePreserveLengthSingleName(t *testing.T) {
	t.Run("single name with preserve length", func(t *testing.T) {
		randomizer := rng.New(1)
		singleName := "John"
		expectedLength := int64(len(singleName))

		// Debug: Let's see what splitEvenly returns
		firstname, lastname := splitEvenly(singleName)
		t.Logf("splitEvenly('%s') = firstname: '%s' (len: %d), lastname: '%s' (len: %d)",
			singleName, firstname, len(firstname), lastname, len(lastname))

		res, err := transformFullName(randomizer, singleName, true, maxCharacterLimit)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		t.Logf("Input: '%s' (len: %d), Output: '%s' (len: %d)",
			singleName, len(singleName), *res, len(*res))
		assert.Equal(t, expectedLength, int64(len(*res)), "The output should be the same length as the input")
		assert.Contains(t, *res, " ", "Should contain a space between first and last name")
	})

	t.Run("single name with preserve length via bloblang", func(t *testing.T) {
		singleName := "John"
		mapping := fmt.Sprintf(`root = transform_full_name(value:%q,preserve_length:true,max_length:%d)`, singleName, maxCharacterLimit)
		ex, err := bloblang.Parse(mapping)
		assert.NoError(t, err, "failed to parse the full name transformer")

		res, err := ex.Query(singleName)
		assert.NoError(t, err)

		resStr, ok := res.(*string)
		if !ok {
			t.Errorf("Expected *string, got %T", res)
			return
		}

		assert.NotNil(t, resStr)
		t.Logf("Bloblang Input: '%s' (len: %d), Output: '%s' (len: %d)",
			singleName, len(singleName), *resStr, len(*resStr))
		assert.Equal(t, len(singleName), len(*resStr), "Generated full name must be as long as input name")
		assert.Contains(t, *resStr, " ", "Should contain a space between first and last name")
	})
}

func Test_TransformFullNamePreserveLengthShortLastName(t *testing.T) {
	t.Run("name with short last name", func(t *testing.T) {
		randomizer := rng.New(1)
		nameWithShortLastName := "John A"
		expectedLength := int64(len(nameWithShortLastName))

		res, err := transformFullName(randomizer, nameWithShortLastName, true, maxCharacterLimit)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, expectedLength, int64(len(*res)), "The output should be the same length as the input")
		assert.Contains(t, *res, " ", "Should contain a space between first and last name")
	})
}

func Test_TransformFullNamePreserveLengthMultipleWords(t *testing.T) {
	t.Run("name with multiple words", func(t *testing.T) {
		randomizer := rng.New(1)
		multiWordName := "John van der Berg"
		expectedLength := int64(len(multiWordName))

		res, err := transformFullName(randomizer, multiWordName, true, maxCharacterLimit)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, expectedLength, int64(len(*res)), "The output should be the same length as the input")
		assert.Contains(t, *res, " ", "Should contain spaces between name parts")
	})
}

func Test_TransformFullNamePreserveLengthEdgeCases(t *testing.T) {
	t.Run("very short name", func(t *testing.T) {
		randomizer := rng.New(1)
		shortName := "A"
		expectedLength := int64(len(shortName))

		res, err := transformFullName(randomizer, shortName, true, maxCharacterLimit)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, expectedLength, int64(len(*res)), "The output should be the same length as the input")
	})

	t.Run("name with only spaces", func(t *testing.T) {
		randomizer := rng.New(1)
		spaceName := "   "

		res, err := transformFullName(randomizer, spaceName, true, maxCharacterLimit)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(len(spaceName)), int64(len(*res)), "The output should be the same length as the input")
	})
}

func Test_SplitEvenly(t *testing.T) {
	t.Run("even number of words", func(t *testing.T) {
		first, last := splitEvenly("John Smith")
		assert.Equal(t, "John", first)
		assert.Equal(t, "Smith", last)
	})

	t.Run("odd number of words", func(t *testing.T) {
		first, last := splitEvenly("John van der Berg")
		assert.Equal(t, "John van", first)
		assert.Equal(t, "der Berg", last)
	})

	t.Run("single word", func(t *testing.T) {
		first, last := splitEvenly("John")
		assert.Equal(t, "John", first)
		assert.Equal(t, "", last)
	})

	t.Run("empty string", func(t *testing.T) {
		first, last := splitEvenly("")
		assert.Equal(t, "", first)
		assert.Equal(t, "", last)
	})

	t.Run("multiple spaces", func(t *testing.T) {
		first, last := splitEvenly("John  Smith")
		assert.Equal(t, "John", first)
		assert.Equal(t, "Smith", last)
	})
}

func Test_TransformFullNamePreserveLength_FirstNameFallback(t *testing.T) {
	t.Run("first name fallback to random string padding", func(t *testing.T) {
		randomizer := rng.New(1)
		// Use a name with a length that is unlikely to exist in the corpus (e.g., 1 character)
		name := "A B"
		res, err := transformFullName(randomizer, name, true, maxCharacterLimit)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, len(name), len(*res), "Output should match input length")
		assert.Contains(t, *res, " ", "Should contain a space")
		// The first and last name parts should each be 1 character (plus the space)
		parts := strings.SplitN(*res, " ", 2)
		if len(parts) == 2 {
			assert.Len(t, parts[0], 1, "First name part should be 1 char (padded if needed)")
			assert.Len(t, parts[1], 1, "Last name part should be 1 char (padded if needed)")
		}
	})
}

func Test_TransformFullNamePreserveLength_PadAndTruncate(t *testing.T) {
	t.Run("padding when generated name is shorter than input", func(t *testing.T) {
		randomizer := rng.New(1)
		// Use a name with a length that is longer than typical generated names
		name := "John Smith ExtraLongName"
		res, err := transformFullName(randomizer, name, true, maxCharacterLimit)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, len(name), len(*res), "Output should match input length")
		assert.Contains(t, *res, " ", "Should contain a space")
	})

	t.Run("truncation when generated name is longer than input", func(t *testing.T) {
		randomizer := rng.New(1)
		// Use a very short name so the generated name will be truncated
		name := "Jo Sm"
		res, err := transformFullName(randomizer, name, true, maxCharacterLimit)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, len(name), len(*res), "Output should match input length")
		assert.Contains(t, *res, " ", "Should contain a space")
	})
}
