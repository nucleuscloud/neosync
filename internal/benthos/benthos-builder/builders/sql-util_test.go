package benthosbuilder_builders

import "testing"

func Test_cleanPostgresType(t *testing.T) {
	t.Run("simple type without params", func(t *testing.T) {
		result := cleanPostgresType("integer")
		expected := "integer"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("type with single parameter", func(t *testing.T) {
		result := cleanPostgresType("varchar(255)")
		expected := "varchar"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("type with multiple parameters", func(t *testing.T) {
		result := cleanPostgresType("numeric(10,2)")
		expected := "numeric"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("character varying with parameter", func(t *testing.T) {
		result := cleanPostgresType("character varying(50)")
		expected := "character varying"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("type with spaces and parameter", func(t *testing.T) {
		result := cleanPostgresType("bit varying(8)")
		expected := "bit varying"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("array type", func(t *testing.T) {
		result := cleanPostgresType("integer[]")
		expected := "integer[]"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("type with time zone", func(t *testing.T) {
		result := cleanPostgresType("timestamp with time zone")
		expected := "timestamp with time zone"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("type with trailing space", func(t *testing.T) {
		result := cleanPostgresType("numeric(10,2) ")
		expected := "numeric"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("type with leading space", func(t *testing.T) {
		result := cleanPostgresType(" decimal(10,2)")
		expected := "decimal"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		result := cleanPostgresType("")
		expected := ""
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("just parentheses", func(t *testing.T) {
		result := cleanPostgresType("()")
		expected := ""
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("multiple parentheses", func(t *testing.T) {
		result := cleanPostgresType("foo(bar(baz))")
		expected := "foo"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("json type", func(t *testing.T) {
		result := cleanPostgresType("json")
		expected := "json"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("jsonb type", func(t *testing.T) {
		result := cleanPostgresType("jsonb")
		expected := "jsonb"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})
}

// Benchmark remains the same as it's already using discrete cases
func BenchmarkCleanPostgresType(b *testing.B) {
	inputs := []string{
		"integer",
		"character varying(50)",
		"numeric(10,2)",
		"timestamp with time zone",
		"text[]",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			cleanPostgresType(input)
		}
	}
}
