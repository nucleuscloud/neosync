package postgres

import (
	"unicode"

	gotypeparser "github.com/nucleuscloud/neosync/internal/gotype-parser"
)

// Tokenizes Postgres array string
// "{"cat", "dog", 1}" -> []string{"{", "cat", ",", "dog", ",", "1", "}"}
type tokenizer struct {
	input string
	pos   int
}

func newTokenizer(input string) *tokenizer {
	return &tokenizer{input: input}
}

func (t *tokenizer) hasNext() bool {
	return t.pos < len(t.input)
}

func (t *tokenizer) next() string {
	t.skipWhitespace()

	if !t.hasNext() {
		return ""
	}

	switch t.input[t.pos] {
	case '{', '}', ',':
		t.pos++
		return string(t.input[t.pos-1])
	case '"':
		return t.readQuotedString()
	default:
		return t.readUnquotedValue()
	}
}

// Returns string values
// Important to keep quotes so we maintain string type when parsing
// {"cat"} -> `"cat"`
// {"hey, {name}"} -> `"hey, {name}"“
func (t *tokenizer) readQuotedString() string {
	start := t.pos
	t.pos++ // skip opening quote
	for t.pos < len(t.input) {
		if t.input[t.pos] == '"' && t.input[t.pos-1] != '\\' {
			t.pos++ // inclued closing quote
			return t.input[start:t.pos]
		}
		t.pos++
	}
	return t.input[start:t.pos]
}

// Returns non string values
// {1} -> 1
func (t *tokenizer) readUnquotedValue() string {
	start := t.pos
	for t.pos < len(t.input) && !isDelimiter(t.input[t.pos]) {
		t.pos++
	}
	return t.input[start:t.pos]
}

func (t *tokenizer) skipWhitespace() {
	for t.pos < len(t.input) && unicode.IsSpace(rune(t.input[t.pos])) {
		t.pos++
	}
}

func isDelimiter(c byte) bool {
	return c == '{' || c == '}' || c == ',' || unicode.IsSpace(rune(c))
}

// Parses Postgres array in tokenized form
// []string{"{", "hey, there {name}", ",", "1", "}"} -> []any{"hey, there {name}", 1}
// []string{"{", "{", "1", ",", "20", "}", ",", "{", "33", ",", "4", "}", "}"} -> [][]any{{1,20}, {33,4}}
type PgArrayParser interface {
	Parse(input string) any
}

type ArrayParser struct{}

func NewArrayParser() PgArrayParser {
	return &ArrayParser{}
}

func (p *ArrayParser) Parse(input string) any {
	if input == "" || input == "{}" {
		return []any{}
	}

	tokenizer := newTokenizer(input)
	return p.parseArray(tokenizer)[0]
}

func (p *ArrayParser) parseArray(t *tokenizer) []any {
	result := []any{}
	for t.hasNext() {
		token := t.next()
		switch token {
		case "{":
			result = append(result, p.parseArray(t))
		case "}":
			return result
		case ",":
			continue
		default:
			result = append(result, parseValue(token))
		}
	}
	return result
}

func parseValue(s string) any {
	if s[0] == '"' && s[len(s)-1] == '"' {
		// quoted string
		return s[1 : len(s)-1]
	}
	// unquoted try to convert to int or float
	if n, err := gotypeparser.ParseStringAsNumber(s); err == nil {
		return n
	}
	return s
}
