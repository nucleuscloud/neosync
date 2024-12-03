package connectionmanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSession_Name(t *testing.T) {
	tests := []struct {
		name     string
		session  *Session
		expected string
	}{
		{
			name:     "basic name",
			session:  NewSession("test-session"),
			expected: "test-session",
		},
		{
			name:     "empty name",
			session:  NewSession(""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.session.Name())
		})
	}
}

func TestSession_Group(t *testing.T) {
	tests := []struct {
		name     string
		session  *Session
		expected string
	}{
		{
			name:     "with group",
			session:  NewSession("test", WithSessionGroup("group1")),
			expected: "group1",
		},
		{
			name:     "without group",
			session:  NewSession("test"),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.session.Group())
		})
	}
}

func TestSession_String(t *testing.T) {
	tests := []struct {
		name     string
		session  *Session
		expected string
	}{
		{
			name:     "with group",
			session:  NewSession("test", WithSessionGroup("group1")),
			expected: "group1:test",
		},
		{
			name:     "without group",
			session:  NewSession("test"),
			expected: "test",
		},
		{
			name:     "empty name with group",
			session:  NewSession("", WithSessionGroup("group1")),
			expected: "group1:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.session.String())
		})
	}
}

func TestNewSession(t *testing.T) {
	tests := []struct {
		name        string
		sessionName string
		opts        []SessionOption
		expected    *Session
	}{
		{
			name:        "basic session",
			sessionName: "test",
			opts:        nil,
			expected: &Session{
				name: "test",
				cfg:  &sessionConfig{},
			},
		},
		{
			name:        "session with group",
			sessionName: "test",
			opts:        []SessionOption{WithSessionGroup("group1")},
			expected: &Session{
				name: "test",
				cfg:  &sessionConfig{group: "group1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := NewSession(tt.sessionName, tt.opts...)
			assert.Equal(t, tt.expected.name, session.name)
			assert.Equal(t, tt.expected.cfg.group, session.cfg.group)
		})
	}
}

func TestNewUniqueSession(t *testing.T) {
	t.Run("creates unique sessions", func(t *testing.T) {
		s1 := NewUniqueSession()
		s2 := NewUniqueSession()
		assert.NotEqual(t, s1.Name(), s2.Name())
	})

	t.Run("respects group option", func(t *testing.T) {
		group := "test-group"
		s := NewUniqueSession(WithSessionGroup(group))
		assert.Equal(t, group, s.Group())
	})
}

func TestSessionOption(t *testing.T) {
	t.Run("multiple options", func(t *testing.T) {
		group1 := "group1"
		group2 := "group2"
		s := NewSession("test",
			WithSessionGroup(group1),
			WithSessionGroup(group2))
		assert.Equal(t, group2, s.Group())
	})
}
