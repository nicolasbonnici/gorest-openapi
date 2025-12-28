package openapi

import (
	"testing"
)

func TestGoTypeToOpenAPIType(t *testing.T) {
	tests := []struct {
		name       string
		goType     string
		wantType   string
		wantFormat string
	}{
		// Integer types
		{
			name:       "int maps to integer with int32 format",
			goType:     "int",
			wantType:   "integer",
			wantFormat: "int32",
		},
		{
			name:       "int32 maps to integer with int32 format",
			goType:     "int32",
			wantType:   "integer",
			wantFormat: "int32",
		},
		{
			name:       "int64 maps to integer with int64 format",
			goType:     "int64",
			wantType:   "integer",
			wantFormat: "int64",
		},
		{
			name:       "int16 maps to integer with int32 format",
			goType:     "int16",
			wantType:   "integer",
			wantFormat: "int32",
		},
		// Float types
		{
			name:       "float32 maps to number with float format",
			goType:     "float32",
			wantType:   "number",
			wantFormat: "float",
		},
		{
			name:       "float64 maps to number with double format",
			goType:     "float64",
			wantType:   "number",
			wantFormat: "double",
		},
		// String types
		{
			name:       "string maps to string with no format",
			goType:     "string",
			wantType:   "string",
			wantFormat: "",
		},
		// Boolean types
		{
			name:       "bool maps to boolean with no format",
			goType:     "bool",
			wantType:   "boolean",
			wantFormat: "",
		},
		// Time types
		{
			name:       "time.Time maps to string with date-time format",
			goType:     "time.Time",
			wantType:   "string",
			wantFormat: "date-time",
		},
		// Interface types
		{
			name:       "interface{} maps to object with no format",
			goType:     "interface{}",
			wantType:   "object",
			wantFormat: "",
		},
		// Pointer types (should strip * prefix)
		{
			name:       "pointer to string",
			goType:     "*string",
			wantType:   "string",
			wantFormat: "",
		},
		{
			name:       "pointer to int64",
			goType:     "*int64",
			wantType:   "integer",
			wantFormat: "int64",
		},
		{
			name:       "pointer to time.Time",
			goType:     "*time.Time",
			wantType:   "string",
			wantFormat: "date-time",
		},
		// Unknown types default to string
		{
			name:       "unknown type defaults to string",
			goType:     "CustomType",
			wantType:   "string",
			wantFormat: "",
		},
		{
			name:       "empty type defaults to string",
			goType:     "",
			wantType:   "string",
			wantFormat: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotFormat := goTypeToOpenAPIType(tt.goType)
			if gotType != tt.wantType {
				t.Errorf("goTypeToOpenAPIType(%q) type = %v, want %v", tt.goType, gotType, tt.wantType)
			}
			if gotFormat != tt.wantFormat {
				t.Errorf("goTypeToOpenAPIType(%q) format = %v, want %v", tt.goType, gotFormat, tt.wantFormat)
			}
		})
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		name string
		word string
		want string
	}{
		// Regular pluralization (add 's')
		{
			name: "regular word - cat",
			word: "cat",
			want: "cats",
		},
		{
			name: "regular word - dog",
			word: "dog",
			want: "dogs",
		},
		// Words ending in 'y' preceded by consonant (y -> ies)
		{
			name: "consonant + y -> ies",
			word: "category",
			want: "categories",
		},
		{
			name: "consonant + y -> ies (city)",
			word: "city",
			want: "cities",
		},
		{
			name: "consonant + y -> ies (lady)",
			word: "lady",
			want: "ladies",
		},
		// Words ending in 'y' preceded by vowel (add 's')
		{
			name: "vowel + y -> ys",
			word: "key",
			want: "keys",
		},
		{
			name: "vowel + y -> ys (boy)",
			word: "boy",
			want: "boys",
		},
		{
			name: "vowel + y -> ys (day)",
			word: "day",
			want: "days",
		},
		// Words ending in 'fe' -> 'ves'
		{
			name: "fe -> ves",
			word: "life",
			want: "lives",
		},
		{
			name: "fe -> ves (wife)",
			word: "wife",
			want: "wives",
		},
		{
			name: "fe -> ves (knife)",
			word: "knife",
			want: "knives",
		},
		// Words ending in 'f' -> 'ves'
		{
			name: "f -> ves",
			word: "leaf",
			want: "leaves",
		},
		{
			name: "f -> ves (wolf)",
			word: "wolf",
			want: "wolves",
		},
		// Words ending in 's' -> 'es'
		{
			name: "s -> es",
			word: "class",
			want: "classes",
		},
		{
			name: "s -> es (bus)",
			word: "bus",
			want: "buses",
		},
		// Words ending in 'x' -> 'es'
		{
			name: "x -> es",
			word: "box",
			want: "boxes",
		},
		{
			name: "x -> es (fox)",
			word: "fox",
			want: "foxes",
		},
		// Words ending in 'z' -> 'es'
		{
			name: "z -> es",
			word: "quiz",
			want: "quizes",
		},
		// Words ending in 'ch' -> 'es'
		{
			name: "ch -> es",
			word: "church",
			want: "churches",
		},
		{
			name: "ch -> es (beach)",
			word: "beach",
			want: "beaches",
		},
		// Words ending in 'sh' -> 'es'
		{
			name: "sh -> es",
			word: "dish",
			want: "dishes",
		},
		{
			name: "sh -> es (brush)",
			word: "brush",
			want: "brushes",
		},
		// Edge cases
		{
			name: "single letter",
			word: "a",
			want: "as",
		},
		{
			name: "two letters",
			word: "ox",
			want: "oxes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pluralize(tt.word); got != tt.want {
				t.Errorf("pluralize(%q) = %v, want %v", tt.word, got, tt.want)
			}
		})
	}
}

func TestIsVowel(t *testing.T) {
	tests := []struct {
		name string
		char byte
		want bool
	}{
		{name: "a is vowel", char: 'a', want: true},
		{name: "e is vowel", char: 'e', want: true},
		{name: "i is vowel", char: 'i', want: true},
		{name: "o is vowel", char: 'o', want: true},
		{name: "u is vowel", char: 'u', want: true},
		{name: "b is not vowel", char: 'b', want: false},
		{name: "c is not vowel", char: 'c', want: false},
		{name: "d is not vowel", char: 'd', want: false},
		{name: "z is not vowel", char: 'z', want: false},
		{name: "A (uppercase) is not vowel", char: 'A', want: false},
		{name: "E (uppercase) is not vowel", char: 'E', want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isVowel(tt.char); got != tt.want {
				t.Errorf("isVowel(%q) = %v, want %v", tt.char, got, tt.want)
			}
		})
	}
}
