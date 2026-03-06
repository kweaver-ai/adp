package dagmodel

import "testing"

func TestToEntity_StringJSONToInterface(t *testing.T) {
	type srcStruct struct {
		Value string
	}
	type destStruct struct {
		Value interface{}
	}

	tests := []struct {
		name   string
		input  string
		assert func(t *testing.T, got interface{})
	}{
		{
			name:  "json object string should unmarshal to map",
			input: "{}",
			assert: func(t *testing.T, got interface{}) {
				t.Helper()
				m, ok := got.(map[string]interface{})
				if !ok {
					t.Fatalf("expected map[string]interface{}, got %T (%v)", got, got)
				}
				if len(m) != 0 {
					t.Fatalf("expected empty map, got %v", m)
				}
			},
		},
		{
			name:  "json null string should become nil interface",
			input: "null",
			assert: func(t *testing.T, got interface{}) {
				t.Helper()
				if got != nil {
					t.Fatalf("expected nil, got %T (%v)", got, got)
				}
			},
		},
		{
			name:  "non json string should keep raw string",
			input: "plain-text",
			assert: func(t *testing.T, got interface{}) {
				t.Helper()
				s, ok := got.(string)
				if !ok {
					t.Fatalf("expected string, got %T (%v)", got, got)
				}
				if s != "plain-text" {
					t.Fatalf("expected plain-text, got %q", s)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := &srcStruct{Value: tt.input}
			dest := &destStruct{Value: map[string]interface{}{"old": true}}

			if err := ToEntity(src, dest); err != nil {
				t.Fatalf("ToEntity returned error: %v", err)
			}

			tt.assert(t, dest.Value)
		})
	}
}
