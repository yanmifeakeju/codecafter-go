package resp

import (
	"strings"
	"testing"
)

func TestEncoderEncode(t *testing.T) {
	tests := []struct {
		name  string
		value Value
		want  string
	}{
		{
			name:  "simple string",
			value: SimpleStringValue("OK"),
			want:  "+OK\r\n",
		},
		{
			name:  "error",
			value: ErrorValue("ERR unknown command"),
			want:  "-ERR unknown command\r\n",
		},
		{
			name:  "integer",
			value: IntValue(42),
			want:  ":42\r\n",
		},
		{
			name:  "negative integer",
			value: IntValue(-1),
			want:  ":-1\r\n",
		},
		{
			name:  "bulk string",
			value: BulkStringValue("hello"),
			want:  "$5\r\nhello\r\n",
		},
		{
			name:  "empty bulk string",
			value: BulkStringValue(""),
			want:  "$0\r\n\r\n",
		},
		{
			name:  "null bulk string",
			value: NullBulkString(),
			want:  "$-1\r\n",
		},
		{
			name:  "array",
			value: ArrayValue([]Value{BulkStringValue("ECHO"), BulkStringValue("hey")}),
			want:  "*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n",
		},
		{
			name:  "empty array",
			value: ArrayValue([]Value{}),
			want:  "*0\r\n",
		},
		{
			name:  "null array",
			value: Value{Type: Array, Null: true},
			want:  "*-1\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got strings.Builder
			enc := NewEncoder(&got)

			if err := enc.Encode(tt.value); err != nil {
				t.Fatalf("Encode returned error: %v", err)
			}

			if got.String() != tt.want {
				t.Fatalf("Encode() = %q, want %q", got.String(), tt.want)
			}
		})
	}
}
