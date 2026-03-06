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

func TestToEntity_ConversionCoverage(t *testing.T) {
	type customString string
	type customInt int64
	type payload struct {
		Name string `json:"name"`
	}

	type srcStruct struct {
		Same         string
		NumToStr     int64
		StrToNum     string
		ConvCustom   string
		ValToPtr     int
		PtrToVal     *int
		PtrToPtr     *int
		JSONToStruct string
		JSONToPtr    string
		StrToPtrNum  string
	}

	type destStruct struct {
		Same         string
		NumToStr     customString
		StrToNum     uint64
		ConvCustom   customString
		ValToPtr     *int64
		PtrToVal     int64
		PtrToPtr     *int64
		JSONToStruct payload
		JSONToPtr    *payload
		StrToPtrNum  *uint64
	}

	ptrVal := 12
	ptrVal2 := 88
	src := &srcStruct{
		Same:         "same-value",
		NumToStr:     99,
		StrToNum:     "100",
		ConvCustom:   "custom",
		ValToPtr:     7,
		PtrToVal:     &ptrVal,
		PtrToPtr:     &ptrVal2,
		JSONToStruct: `{"name":"alice"}`,
		JSONToPtr:    `{"name":"bob"}`,
		StrToPtrNum:  "123",
	}

	dest := &destStruct{}
	if err := ToEntity(src, dest); err != nil {
		t.Fatalf("ToEntity returned error: %v", err)
	}

	if dest.Same != "same-value" {
		t.Fatalf("Same mismatch: %q", dest.Same)
	}
	if string(dest.NumToStr) != "99" {
		t.Fatalf("NumToStr mismatch: %q", dest.NumToStr)
	}
	if dest.StrToNum != 100 {
		t.Fatalf("StrToNum mismatch: %d", dest.StrToNum)
	}
	if string(dest.ConvCustom) != "custom" {
		t.Fatalf("ConvCustom mismatch: %q", dest.ConvCustom)
	}
	if dest.ValToPtr == nil || *dest.ValToPtr != 7 {
		t.Fatalf("ValToPtr mismatch: %v", dest.ValToPtr)
	}
	if dest.PtrToVal != 12 {
		t.Fatalf("PtrToVal mismatch: %d", dest.PtrToVal)
	}
	if dest.PtrToPtr == nil || *dest.PtrToPtr != 88 {
		t.Fatalf("PtrToPtr mismatch: %v", dest.PtrToPtr)
	}
	if dest.JSONToStruct.Name != "alice" {
		t.Fatalf("JSONToStruct mismatch: %+v", dest.JSONToStruct)
	}
	if dest.JSONToPtr == nil || dest.JSONToPtr.Name != "bob" {
		t.Fatalf("JSONToPtr mismatch: %+v", dest.JSONToPtr)
	}
	if dest.StrToPtrNum == nil || *dest.StrToPtrNum != 123 {
		t.Fatalf("StrToPtrNum mismatch: %v", dest.StrToPtrNum)
	}

	var _ customInt = customInt(dest.PtrToVal)
}
