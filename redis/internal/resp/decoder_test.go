package resp

import (
	"reflect"
	"strings"
	"testing"
)

func TestDecoderDecode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Value
	}{
		{
			name:  "bulk string",
			input: "$4\r\nECHO\r\n",
			want:  BulkStringValue("ECHO"),
		},
		{
			name:  "empty bulk string",
			input: "$0\r\n\r\n",
			want:  BulkStringValue(""),
		},
		{
			name:  "null bulk string",
			input: "$-1\r\n",
			want:  NullBulkString(),
		},
		{
			name:  "array of bulk strings",
			input: "*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n",
			want:  ArrayValue([]Value{BulkStringValue("ECHO"), BulkStringValue("hey")}),
		},
		{
			name:  "empty array",
			input: "*0\r\n",
			want:  ArrayValue([]Value{}),
		},
		{
			name:  "null array",
			input: "*-1\r\n",
			want:  Value{Type: Array, Null: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := NewDecoder(strings.NewReader(tt.input))

			var got Value
			if err := dec.Decode(&got); err != nil {
				t.Fatalf("Decode returned error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Decode() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestDecoderDecodeErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "unsupported type",
			input: "+OK\r\n",
		},
		{
			name:  "missing CRLF in length line",
			input: "$4\nECHO\r\n",
		},
		{
			name:  "invalid bulk string length",
			input: "$abc\r\n",
		},
		{
			name:  "negative bulk string length",
			input: "$-2\r\n",
		},
		{
			name:  "missing bulk string terminator",
			input: "$4\r\nECHO\n",
		},
		{
			name:  "short bulk string body",
			input: "$4\r\nEC",
		},
		{
			name:  "invalid array length",
			input: "*abc\r\n",
		},
		{
			name:  "negative array length",
			input: "*-2\r\n",
		},
		{
			name:  "short array",
			input: "*2\r\n$4\r\nECHO\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := NewDecoder(strings.NewReader(tt.input))

			var got Value
			if err := dec.Decode(&got); err == nil {
				t.Fatal("Decode returned nil error")
			}
		})
	}
}

func TestDecoderDecodeNilValue(t *testing.T) {
	dec := NewDecoder(strings.NewReader("$4\r\nECHO\r\n"))

	if err := dec.Decode(nil); err == nil {
		t.Fatal("Decode returned nil error")
	}
}
