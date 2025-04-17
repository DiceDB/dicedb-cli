package ironhawk

import (
	"reflect"
	"testing"
)

func TestParseArgs_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Simple command with double quotes",
			input:    `SET key "hello world"`,
			expected: []string{"SET", "key", "hello world"},
		},
		{
			name:     "Escaped quote inside string",
			input:    `ECHO "She said \"hello\""`,
			expected: []string{"ECHO", `She said "hello"`},
		},
		{
			name:     "Single quote inside double quotes",
			input:    `ECHO "It\'s a test"`,
			expected: []string{"ECHO", `It's a test`},
		},
		{
			name:     "Double quote inside single quotes",
			input:    `ECHO 'She said "hi"'`,
			expected: []string{"ECHO", `She said "hi"`},
		},
		{
			name:     "Unterminated quote (should fail)",
			input:    `SET key "unterminated value`,
			expected: []string{"SET", "key", "unterminated value"},
		},
		{
			name:     "Nested quotes",
			input:    `CMD "value with 'nested quote'"`,
			expected: []string{"CMD", `value with 'nested quote'`},
		},
		{
			name:     "Empty quoted argument",
			input:    `SET key ""`,
			expected: []string{"SET", "key", ""},
		},
		{
			name:     "Trailing space",
			input:    `ECHO "hello world" `,
			expected: []string{"ECHO", "hello world"},
		},
		{
			name:     "Special characters in value",
			input:    `SET key "!@#$%^&*()"`,
			expected: []string{"SET", "key", "!@#$%^&*()"},
		},
		{
			name:     "Argument with escaped backslash",
			input:    `ECHO "C:\\Program Files\\App"`,
			expected: []string{"ECHO", `C:\Program Files\App`},
		},
		{
			name:     "Multiple spaces between arguments",
			input:    `SET     key     "hello"`,
			expected: []string{"SET", "key", "hello"},
		},
		{
			name:     "Command with tab separators",
			input:    "SET\tkey\t\"value with tab\"",
			expected: []string{"SET", "key", "value with tab"},
		},
		{
			name:     "List-like input",
			input:    `RPUSH list "item1" "item2" "item3"`,
			expected: []string{"RPUSH", "list", "item1", "item2", "item3"},
		},
		{
			name:     "Command with newline in input",
			input:    "SET key \"value with \n newline\"",
			expected: []string{"SET", "key", "value with \n newline"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseArgs(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("parseArgs(%q) = %#v; expected %#v", tc.input, result, tc.expected)
			}
		})
	}
}
