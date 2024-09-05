package postgres

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ArrayParser(t *testing.T) {
	parser := NewArrayParser()

	t.Run("Empty input", func(t *testing.T) {
		result := parser.Parse("")
		require.Equal(t, []any{}, result)
	})

	t.Run("Empty array", func(t *testing.T) {
		result := parser.Parse("{}")
		require.Equal(t, []any{}, result)
	})

	t.Run("Single-level array mixed", func(t *testing.T) {
		result := parser.Parse(`{1,2, "cat", "dog", 1.2, {3, 4}}`)
		require.Equal(t, []any{int64(1), int64(2), "cat", "dog", 1.2, []any{int64(3), int64(4)}}, result)
	})

	t.Run("Two-level nested array", func(t *testing.T) {
		result := parser.Parse("{{1,2,3}, {4,5}}")
		expected := [][]any{{int64(1), int64(2), int64(3)}, {int64(4), int64(5)}}
		require.Len(t, result, len(expected))
		resultArray, ok := result.([]any)
		require.True(t, ok)
		for i, a := range resultArray {
			require.Equal(t, expected[i], a)
		}
	})

	t.Run("Three-level nested array", func(t *testing.T) {
		result := parser.Parse("{{{1,2},{4,5}},{{8,9}}}")
		fmt.Println(result)
		expected := [][][]any{{{int64(1), int64(2)}, {int64(4), int64(5)}}, {{int64(8), int64(9)}}}
		require.Len(t, result, len(expected))
		resultArray, ok := result.([]any)
		require.True(t, ok)
		for i, aa := range resultArray {
			innerArray, ok := aa.([]any)
			require.True(t, ok)
			for j, a := range innerArray {
				require.Equal(t, expected[i][j], a)
			}
		}
	})

	t.Run("Array with spaces", func(t *testing.T) {
		result := parser.Parse(" { 1 , 2 , 3 } ")
		fmt.Println(result)
		require.Equal(t, []any{int64(1), int64(2), int64(3)}, result)
	})
}

func Test_Tokenizer(t *testing.T) {
	t.Run("Empty input", func(t *testing.T) {
		tokenizer := newTokenizer("")
		var tokens []string
		for tokenizer.hasNext() {
			tokens = append(tokens, tokenizer.next())
		}
		require.Empty(t, tokens)
	})

	t.Run("Simple array", func(t *testing.T) {
		tokenizer := newTokenizer("{1,2,3}")
		var tokens []string
		for tokenizer.hasNext() {
			tokens = append(tokens, tokenizer.next())
		}
		require.Equal(t, []string{"{", "1", ",", "2", ",", "3", "}"}, tokens)
	})

	t.Run("Nested array", func(t *testing.T) {
		tokenizer := newTokenizer("{{1,20, 2.0},{33,4}}")
		var tokens []string
		for tokenizer.hasNext() {
			tokens = append(tokens, tokenizer.next())
		}
		require.Equal(t, []string{"{", "{", "1", ",", "20", ",", "2.0", "}", ",", "{", "33", ",", "4", "}", "}"}, tokens)
	})

	t.Run("Array with strings", func(t *testing.T) {
		tokenizer := newTokenizer(`{"cat","lizard", 1.2}`)
		var tokens []string
		for tokenizer.hasNext() {
			tokens = append(tokens, tokenizer.next())
		}
		require.Equal(t, []string{"{", `"cat"`, ",", `"lizard"`, ",", "1,2", "}"}, tokens)
	})

	t.Run("Array with strings", func(t *testing.T) {
		tokenizer := newTokenizer(`{"hey, there {name}","lizard"}`)
		var tokens []string
		for tokenizer.hasNext() {
			tokens = append(tokens, tokenizer.next())
		}
		require.Equal(t, []string{"{", `"hey, there {name}"`, ",", `"lizard"`, "}"}, tokens)
	})

	t.Run("Array with spaces", func(t *testing.T) {
		tokenizer := newTokenizer(" { 1 , 2, 3 } ")
		var tokens []string
		for tokenizer.hasNext() {
			token := tokenizer.next()
			if token != "" {
				tokens = append(tokens, token)
			}
		}
		require.Equal(t, []string{"{", "1", ",", "2", ",", "3", "}"}, tokens)
	})
}
