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

func TestParseLength(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    int64
		wantErr bool
	}{
		{
			name:  "zero",
			input: []byte("0"),
			want:  0,
		},
		{
			name:  "positive",
			input: []byte("123"),
			want:  123,
		},
		{
			name:  "negative",
			input: []byte("-1"),
			want:  -1,
		},
		{
			name:    "empty",
			input:   []byte(""),
			wantErr: true,
		},
		{
			name:    "missing digits after negative sign",
			input:   []byte("-"),
			wantErr: true,
		},
		{
			name:    "invalid digit",
			input:   []byte("12a"),
			wantErr: true,
		},
		{
			name:    "positive overflow",
			input:   []byte("9223372036854775808"),
			wantErr: true,
		},
		{
			name:  "minimum int64",
			input: []byte("-9223372036854775808"),
			want:  -9223372036854775808,
		},
		{
			name:  "max int64",
			input: []byte("9223372036854775807"),
			want:  9223372036854775807,
		},
		{
			name:    "negative overflow",
			input:   []byte("-9223372036854775809"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLength(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("parseLength returned nil error")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseLength returned error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("parseLength() = %d, want %d", got, tt.want)
			}
		})
	}
}
