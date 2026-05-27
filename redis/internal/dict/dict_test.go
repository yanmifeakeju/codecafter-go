package dict_test

import (
	"reflect"
	"testing"

	"github.com/yanmifeakeju/codecafter-go/redis/internal/dict"
)

func TestDict(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "set then get",
			test: func(t *testing.T) {
				d := dict.New()
				want := dict.Entry{Value: dict.Value{Kind: dict.KindString, Data: "bar"}}
				d.Set("foo", want)

				got, ok := d.Get("foo")
				if !ok {
					t.Fatal("Get returned ok=false")
				}

				if !reflect.DeepEqual(got, want) {
					t.Fatalf("Get() = %#v, want %#v", got, want)
				}
			},
		},
		{
			name: "set, get, delete and get",
			test: func(t *testing.T) {
				d := dict.New()
				want := dict.Entry{Value: dict.Value{Kind: dict.KindList, Data: []string{"test"}}}

				d.Set("foo", want)

				if _, ok := d.Get("foo"); !ok {
					t.Fatal("Get returned ok=false")
				}

				if ok := d.Delete("foo"); !ok {
					t.Fatal("Delete returned ok=false")
				}

				if _, ok := d.Get("foo"); ok {
					t.Fatal("Get returned ok=true after Delete")
				}

			},
		},
		{
			name: "get missing key",
			test: func(t *testing.T) {
				d := dict.New()

				if _, ok := d.Get("foo"); ok {
					t.Fatal("Get returned ok=true")
				}
			},
		},
		{
			name: "delete missing key",
			test: func(t *testing.T) {
				d := dict.New()

				if ok := d.Delete("foo"); ok {
					t.Fatal("Delete returned ok=true")
				}
			},
		},
		{
			name: "set existing key overwrites old entry",
			test: func(t *testing.T) {
				d := dict.New()

				want := dict.Entry{Value: dict.Value{Kind: dict.KindList, Data: []string{"test"}}}
				d.Set("foo", want)

				got, ok := d.Get("foo")
				if !ok {
					t.Fatal("Get returned ok=false")
				}

				if !reflect.DeepEqual(got, want) {
					t.Fatalf("Get() = %#v, want %#v", got, want)
				}

				want = dict.Entry{Value: dict.Value{Kind: dict.KindString, Data: "test"}}
				d.Set("foo", want)

				got, ok = d.Get("foo")
				if !ok {
					t.Fatal("Get returned ok=false")
				}

				if !reflect.DeepEqual(got, want) {
					t.Fatalf("Get() = %#v, want %#v", got, want)
				}

			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}

}

func TestKindString(t *testing.T) {
	tests := []struct {
		name string
		kind dict.Kind
		want string
	}{
		{
			name: "unknown",
			kind: dict.KindUnknown,
			want: "unknown",
		},
		{
			name: "string",
			kind: dict.KindString,
			want: "string",
		},
		{
			name: "list",
			kind: dict.KindList,
			want: "list",
		},
		{
			name: "unrecognized",
			kind: dict.Kind(99),
			want: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.want {
				t.Fatalf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
