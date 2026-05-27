package store_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/yanmifeakeju/codecafter-go/redis/internal/dict"
	"github.com/yanmifeakeju/codecafter-go/redis/internal/list"
	"github.com/yanmifeakeju/codecafter-go/redis/internal/store"
)

func TestStoreString(t *testing.T) {
	testCases := []struct {
		name string
		dict *dict.Dict
		test func(t *testing.T, s *store.Store)
	}{
		{
			name: "set key and read",
			dict: nil,
			test: func(t *testing.T, s *store.Store) {
				key := "foo"
				value := "bar"
				s.SetString(key, value, time.Time{})

				v, ok, err := s.GetString(key)

				if err != nil {
					t.Fatalf("GetString() error = %v, want nil", err)
				}

				if !ok {
					t.Fatalf("GetString() ok = false, want true")
				}

				if v != value {
					t.Fatalf("GetString() = %q, want %q", v, value)
				}
			},
		},
		{
			name: "get missing key",
			dict: nil,
			test: func(t *testing.T, s *store.Store) {
				key := "foo"

				v, ok, err := s.GetString(key)

				if err != nil {
					t.Fatalf("GetString() error = %v, want nil", err)
				}

				if ok {
					t.Fatalf("GetString() ok = true, want false")
				}

				if v != "" {
					t.Fatalf("GetString() = %q, want %q", v, "")
				}
			},
		},
		{
			name: "get wrong type for key",
			dict: dictWithValue("foo", dict.Value{Kind: dict.KindList, Data: list.New()}),
			test: func(t *testing.T, s *store.Store) {
				key := "foo"

				v, ok, err := s.GetString(key)

				if !errors.Is(err, store.ErrWrongType) {
					t.Fatalf("GetString() error = %v, want %v", err, store.ErrWrongType)
				}

				if ok {
					t.Fatalf("GetString() ok = true, want false")
				}

				if v != "" {
					t.Fatalf("GetString() = %q, want %q", v, "")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := store.New(tc.dict)
			tc.test(t, s)
		})
	}
}

func TestStoreListPushRange(t *testing.T) {
	testCases := []struct {
		name string
		dict *dict.Dict
		test func(t *testing.T, s *store.Store)
	}{
		{
			name: "list operations on non-list",
			dict: dictWithValue("foo", dict.Value{Kind: dict.KindString, Data: "hello"}),
			test: func(t *testing.T, s *store.Store) {
				key := "foo"

				// rpush
				_, err := s.RPush(key, "bar1", "bar2")
				if !errors.Is(err, store.ErrWrongType) {
					t.Fatalf("RPush() error = %v, want %v", err, store.ErrWrongType)
				}

				// lpush
				_, err = s.LPush(key, "bar1", "bar2")
				if !errors.Is(err, store.ErrWrongType) {
					t.Fatalf("LPush() error = %v, want %v", err, store.ErrWrongType)
				}

				// lrange
				_, err = s.LRange(key, 0, -1)
				if !errors.Is(err, store.ErrWrongType) {
					t.Fatalf("LRange() error = %v, want %v", err, store.ErrWrongType)
				}

				// llen
				_, err = s.LLen(key)
				if !errors.Is(err, store.ErrWrongType) {
					t.Fatalf("LLen() error = %v, want %v", err, store.ErrWrongType)
				}

				v, ok, err := s.GetString(key)
				if err != nil {
					t.Fatalf("GetString() error = %v, want nil", err)
				}
				if !ok || v != "hello" {
					t.Fatalf("GetString() = %q, %v; want %q, true", v, ok, "hello")
				}
			},
		},
		{
			name: "rpush creates list",
			dict: nil,
			test: func(t *testing.T, s *store.Store) {
				key := "foo"

				n, err := s.RPush(key, "bar1", "bar2")

				if err != nil {
					t.Fatalf("RPush() error = %v, want nil", err)
				}

				ln, err := s.LLen(key)
				if err != nil {
					t.Fatalf("LLen() error = %v, want nil", err)
				}

				if n != 2 {
					t.Fatalf("RPush() = %d, want 2", n)
				}

				if ln != 2 {
					t.Fatalf("LLen() = %d, want 2", ln)
				}

				items, err := s.LRange(key, 0, -1)

				if err != nil {
					t.Fatalf("LRange() error = %v, want nil", err)
				}

				want := []string{"bar1", "bar2"}
				if !reflect.DeepEqual(items, want) {
					t.Fatalf("LRange() = %#v, want %#v", items, want)
				}
			},
		},
		{
			name: "rpush appends to existing list",
			dict: dictWithValue("foo", dict.Value{Kind: dict.KindList, Data: list.New("bar")}),
			test: func(t *testing.T, s *store.Store) {
				key := "foo"

				n, err := s.RPush(key, "bar1", "bar2")

				if err != nil {
					t.Fatalf("RPush() error = %v, want nil", err)
				}

				if n != 3 {
					t.Fatalf("RPush() = %d, want 3", n)
				}

				ln, err := s.LLen(key)
				if err != nil {
					t.Fatalf("LLen() error = %v, want nil", err)
				}

				if ln != 3 {
					t.Fatalf("LLen() = %d, want 3", ln)
				}

				items, err := s.LRange(key, 0, -1)

				if err != nil {
					t.Fatalf("LRange() error = %v, want nil", err)
				}

				want := []string{"bar", "bar1", "bar2"}
				if !reflect.DeepEqual(items, want) {
					t.Fatalf("LRange() = %#v, want %#v", items, want)
				}
			},
		},
		{
			name: "lpush creates list",
			dict: nil,
			test: func(t *testing.T, s *store.Store) {
				key := "foo"

				n, err := s.LPush(key, "bar1", "bar2")

				if err != nil {
					t.Fatalf("LPush() error = %v, want nil", err)
				}

				ln, err := s.LLen(key)
				if err != nil {
					t.Fatalf("LLen() error = %v, want nil", err)
				}

				if n != 2 {
					t.Fatalf("LPush() = %d, want 2", n)
				}

				if ln != 2 {
					t.Fatalf("LLen() = %d, want 2", ln)
				}

				items, err := s.LRange(key, 0, -1)

				if err != nil {
					t.Fatalf("LRange() error = %v, want nil", err)
				}

				want := []string{"bar2", "bar1"}
				if !reflect.DeepEqual(items, want) {
					t.Fatalf("LRange() = %#v, want %#v", items, want)
				}
			},
		},
		{
			name: "lpush prepends to existing list",
			dict: dictWithValue("foo", dict.Value{Kind: dict.KindList, Data: list.New("bar")}),
			test: func(t *testing.T, s *store.Store) {
				key := "foo"

				n, err := s.LPush(key, "bar1", "bar2")

				if err != nil {
					t.Fatalf("LPush() error = %v, want nil", err)
				}

				ln, err := s.LLen(key)
				if err != nil {
					t.Fatalf("LLen() error = %v, want nil", err)
				}

				if n != 3 {
					t.Fatalf("LPush() = %d, want 3", n)
				}

				if ln != 3 {
					t.Fatalf("LLen() = %d, want 3", ln)
				}

				items, err := s.LRange(key, 0, -1)

				if err != nil {
					t.Fatalf("LRange() error = %v, want nil", err)
				}

				want := []string{"bar2", "bar1", "bar"}
				if !reflect.DeepEqual(items, want) {
					t.Fatalf("LRange() = %#v, want %#v", items, want)
				}
			},
		},
		{
			name: "range and len on missing key",
			dict: nil,
			test: func(t *testing.T, s *store.Store) {
				i, err := s.LRange("foo", 0, -1)
				if err != nil {
					t.Fatalf("LRange() error = %v, want nil", err)
				}

				if len(i) != 0 {
					t.Fatalf("LRange() length = %d, want 0", len(i))
				}

				n, err := s.LLen("foo")
				if err != nil {
					t.Fatalf("LLen() error = %v, want nil", err)
				}

				if n != 0 {
					t.Fatalf("LLen() n = %d, want 0", n)
				}
			},
		},
		{
			name: "lrange normal bound",
			dict: dictWithValue("foo", dict.Value{Kind: dict.KindList, Data: list.New()}),
			test: func(t *testing.T, s *store.Store) {
				n, err := s.RPush("foo", "a", "b", "c")

				if err != nil {
					t.Fatalf("RPush() error = %v, want nil", err)
				}

				if n != 3 {
					t.Fatalf("RPush() = %d, want 3", n)
				}

				items, err := s.LRange("foo", 0, 1)
				if err != nil {
					t.Fatalf("LRange() error = %v, want nil", err)
				}

				want := []string{"a", "b"}
				if !reflect.DeepEqual(items, want) {
					t.Fatalf("LRange() = %#v, want %#v", items, want)
				}
			},
		},
		{
			name: "lrange negative bound",
			dict: dictWithValue("foo", dict.Value{Kind: dict.KindList, Data: list.New()}),
			test: func(t *testing.T, s *store.Store) {
				n, err := s.RPush("foo", "a", "b", "c")

				if err != nil {
					t.Fatalf("RPush() error = %v, want nil", err)
				}

				if n != 3 {
					t.Fatalf("RPush() = %d, want 3", n)
				}

				items, err := s.LRange("foo", -2, -1)
				if err != nil {
					t.Fatalf("LRange() error = %v, want nil", err)
				}

				want := []string{"b", "c"}
				if !reflect.DeepEqual(items, want) {
					t.Fatalf("LRange() = %#v, want %#v", items, want)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := store.New(tc.dict)
			tc.test(t, s)
		})
	}
}

func TestStoreListPop(t *testing.T) {
	testCases := []struct {
		name string
		dict *dict.Dict
		test func(t *testing.T, s *store.Store)
	}{
		{
			name: "pop operations on a non-list",
			dict: dictWithValue("foo", dict.Value{Kind: dict.KindString, Data: "hello"}),
			test: func(t *testing.T, s *store.Store) {
				_, ok, err := s.LPop("foo")
				if !errors.Is(err, store.ErrWrongType) {
					t.Fatalf("LPop() error = %v, want %v", err, store.ErrWrongType)
				}

				if ok {
					t.Fatalf("LPop() ok = true, want false")
				}

				i, err := s.LPopN("foo", 1)
				if !errors.Is(err, store.ErrWrongType) {
					t.Fatalf("LPopN() error = %v, want %v", err, store.ErrWrongType)
				}

				if i != nil {
					t.Fatalf("LPopN() = %#v, want nil", i)
				}

				_, ok, err = s.BLPop("foo", time.Millisecond)
				if !errors.Is(err, store.ErrWrongType) {
					t.Fatalf("BLPop() error = %v, want %v", err, store.ErrWrongType)
				}

				if ok {
					t.Fatalf("BLPop() ok = true, want false")
				}
			},
		},
		{
			name: "pop operations on missing key",
			dict: nil,
			test: func(t *testing.T, s *store.Store) {
				v, ok, err := s.LPop("foo")

				if err != nil {
					t.Fatalf("LPop() error = %v, want nil", err)
				}

				if ok {
					t.Fatalf("LPop() ok = true, want false")
				}

				if v != "" {
					t.Fatalf("LPop() = %q, want %q", v, "")
				}

				items, err := s.LPopN("foo", 1)
				if err != nil {
					t.Fatalf("LPopN() error = %v, want nil", err)
				}

				if len(items) != 0 {
					t.Fatalf("LPopN() length = %d, want 0", len(items))
				}
			},
		},
		{
			name: "lpop on existing list",
			dict: dictWithValue("foo", dict.Value{Kind: dict.KindList, Data: list.New("a", "b", "c")}),
			test: func(t *testing.T, s *store.Store) {
				v, ok, err := s.LPop("foo")

				if err != nil {
					t.Fatalf("LPop() error = %v, want nil", err)
				}

				if !ok {
					t.Fatalf("LPop() ok = false, want true")
				}

				if v != "a" {
					t.Fatalf("LPop() = %q, want %q", v, "a")
				}

				ln, err := s.LLen("foo")
				if err != nil {
					t.Fatalf("LLen() error = %v, want nil", err)
				}

				if ln != 2 {
					t.Fatalf("LLen() = %d, want 2", ln)
				}

				items, err := s.LRange("foo", 0, -1)
				if err != nil {
					t.Fatalf("LRange() error = %v, want nil", err)
				}

				want := []string{"b", "c"}
				if !reflect.DeepEqual(items, want) {
					t.Fatalf("LRange() = %#v, want %#v", items, want)
				}
			},
		},
		{
			name: "lpopn count less than length",
			dict: dictWithValue("foo", dict.Value{Kind: dict.KindList, Data: list.New("a", "b", "c", "d")}),
			test: func(t *testing.T, s *store.Store) {
				items, err := s.LPopN("foo", 2)

				if err != nil {
					t.Fatalf("LPopN() error = %v, want nil", err)
				}

				if len(items) != 2 {
					t.Fatalf("LPopN() length = %d, want 2", len(items))
				}

				want := []string{"a", "b"}
				if !reflect.DeepEqual(items, want) {
					t.Fatalf("LPopN() = %#v, want %#v", items, want)
				}

				items, err = s.LRange("foo", 0, -1)
				if err != nil {
					t.Fatalf("LRange() error = %v, want nil", err)
				}

				want = []string{"c", "d"}
				if !reflect.DeepEqual(items, want) {
					t.Fatalf("LRange() = %#v, want %#v", items, want)
				}
			},
		},
		{
			name: "lpopn count greater than length",
			dict: dictWithValue("foo", dict.Value{Kind: dict.KindList, Data: list.New("a", "b", "c", "d")}),
			test: func(t *testing.T, s *store.Store) {
				items, err := s.LPopN("foo", 10)

				if err != nil {
					t.Fatalf("LPopN() error = %v, want nil", err)
				}

				if len(items) != 4 {
					t.Fatalf("LPopN() length = %d, want 4", len(items))
				}

				want := []string{"a", "b", "c", "d"}
				if !reflect.DeepEqual(items, want) {
					t.Fatalf("LPopN() = %#v, want %#v", items, want)
				}

				items, err = s.LRange("foo", 0, -1)
				if err != nil {
					t.Fatalf("LRange() error = %v, want nil", err)
				}

				if len(items) != 0 {
					t.Fatalf("LRange() length = %d, want 0", len(items))
				}
			},
		},
		{
			name: "lpopn zero count",
			dict: dictWithValue("foo", dict.Value{Kind: dict.KindList, Data: list.New("a", "b", "c")}),
			test: func(t *testing.T, s *store.Store) {
				items, err := s.LPopN("foo", 0)

				if err != nil {
					t.Fatalf("LPopN() error = %v, want nil", err)
				}

				if len(items) != 0 {
					t.Fatalf("LPopN() length = %d, want 0", len(items))
				}

				items, err = s.LRange("foo", 0, -1)
				if err != nil {
					t.Fatalf("LRange() error = %v, want nil", err)
				}

				want := []string{"a", "b", "c"}
				if !reflect.DeepEqual(items, want) {
					t.Fatalf("LRange() = %#v, want %#v", items, want)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := store.New(tc.dict)
			tc.test(t, s)
		})
	}
}

func TestStoreReadAndPushExpired(t *testing.T) {
	expiredAt := time.Now().Add(-time.Second)

	testCases := []struct {
		name string
		dict *dict.Dict
		test func(t *testing.T, s *store.Store)
	}{
		{
			name: "get string treats expired key as missing",
			dict: dictWithEntry("foo", dict.Entry{
				Value:     dict.Value{Kind: dict.KindString, Data: "hello"},
				ExpiresAt: expiredAt,
			}),
			test: func(t *testing.T, s *store.Store) {
				v, ok, err := s.GetString("foo")

				if err != nil {
					t.Fatalf("GetString() error = %v, want nil", err)
				}

				if ok {
					t.Fatalf("GetString() ok = true, want false")
				}

				if v != "" {
					t.Fatalf("GetString() = %q, want %q", v, "")
				}
			},
		},
		{
			name: "range and len treat expired list as missing",
			dict: dictWithEntry("foo", dict.Entry{
				Value:     dict.Value{Kind: dict.KindList, Data: list.New("a", "b")},
				ExpiresAt: expiredAt,
			}),
			test: func(t *testing.T, s *store.Store) {
				items, err := s.LRange("foo", 0, -1)
				if err != nil {
					t.Fatalf("LRange() error = %v, want nil", err)
				}

				if len(items) != 0 {
					t.Fatalf("LRange() length = %d, want 0", len(items))
				}

				n, err := s.LLen("foo")
				if err != nil {
					t.Fatalf("LLen() error = %v, want nil", err)
				}

				if n != 0 {
					t.Fatalf("LLen() = %d, want 0", n)
				}
			},
		},
		{
			name: "pop treats expired list as missing",
			dict: dictWithEntry("foo", dict.Entry{
				Value:     dict.Value{Kind: dict.KindList, Data: list.New("a", "b")},
				ExpiresAt: expiredAt,
			}),
			test: func(t *testing.T, s *store.Store) {
				v, ok, err := s.LPop("foo")
				if err != nil {
					t.Fatalf("LPop() error = %v, want nil", err)
				}

				if ok {
					t.Fatalf("LPop() ok = true, want false")
				}

				if v != "" {
					t.Fatalf("LPop() = %q, want %q", v, "")
				}

				items, err := s.LPopN("foo", 1)
				if err != nil {
					t.Fatalf("LPopN() error = %v, want nil", err)
				}

				if len(items) != 0 {
					t.Fatalf("LPopN() length = %d, want 0", len(items))
				}
			},
		},
		{
			name: "blocking pop treats expired list as missing",
			dict: dictWithEntry("foo", dict.Entry{
				Value:     dict.Value{Kind: dict.KindList, Data: list.New("a")},
				ExpiresAt: expiredAt,
			}),
			test: func(t *testing.T, s *store.Store) {
				v, ok, err := s.BLPop("foo", time.Millisecond)
				if err != nil {
					t.Fatalf("BLPop() error = %v, want nil", err)
				}

				if ok {
					t.Fatalf("BLPop() ok = true, want false")
				}

				if v != "" {
					t.Fatalf("BLPop() = %q, want %q", v, "")
				}
			},
		},
		{
			name: "rpush treats expired string as missing",
			dict: dictWithEntry("foo", dict.Entry{
				Value:     dict.Value{Kind: dict.KindString, Data: "hello"},
				ExpiresAt: expiredAt,
			}),
			test: func(t *testing.T, s *store.Store) {
				n, err := s.RPush("foo", "a", "b")
				if err != nil {
					t.Fatalf("RPush() error = %v, want nil", err)
				}

				if n != 2 {
					t.Fatalf("RPush() = %d, want 2", n)
				}

				items, err := s.LRange("foo", 0, -1)
				if err != nil {
					t.Fatalf("LRange() error = %v, want nil", err)
				}

				want := []string{"a", "b"}
				if !reflect.DeepEqual(items, want) {
					t.Fatalf("LRange() = %#v, want %#v", items, want)
				}
			},
		},
		{
			name: "lpush treats expired string as missing",
			dict: dictWithEntry("foo", dict.Entry{
				Value:     dict.Value{Kind: dict.KindString, Data: "hello"},
				ExpiresAt: expiredAt,
			}),
			test: func(t *testing.T, s *store.Store) {
				n, err := s.LPush("foo", "a", "b")
				if err != nil {
					t.Fatalf("LPush() error = %v, want nil", err)
				}

				if n != 2 {
					t.Fatalf("LPush() = %d, want 2", n)
				}

				items, err := s.LRange("foo", 0, -1)
				if err != nil {
					t.Fatalf("LRange() error = %v, want nil", err)
				}

				want := []string{"b", "a"}
				if !reflect.DeepEqual(items, want) {
					t.Fatalf("LRange() = %#v, want %#v", items, want)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := store.New(tc.dict)
			tc.test(t, s)
		})
	}
}

func dictWithValue(key string, value dict.Value) *dict.Dict {
	return dictWithEntry(key, dict.Entry{Value: value})
}

func dictWithEntry(key string, entry dict.Entry) *dict.Dict {
	d := dict.New()
	d.Set(key, entry)
	return d
}
