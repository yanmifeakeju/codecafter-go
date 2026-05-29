package store_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/yanmifeakeju/codecafter-go/redis/internal/store"
)

type storeTestCase struct {
	name  string
	setup func(*store.Store) error
	test  func(t *testing.T, s *store.Store)
}

type blpopResult struct {
	value string
	ok    bool
	err   error
}

func TestStoreString(t *testing.T) {
	testCases := []storeTestCase{
		{
			name: "set key and read",
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
			name:  "get wrong type for key",
			setup: setupList("foo", "bar"),
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
	runStoreTests(t, testCases)
}

func TestStoreListPushRange(t *testing.T) {
	testCases := []storeTestCase{
		{
			name:  "list operations on non-list",
			setup: setupString("foo", "hello", time.Time{}),
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
			name:  "rpush appends to existing list",
			setup: setupList("foo", "bar"),
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
			name:  "lpush prepends to existing list",
			setup: setupList("foo", "bar"),
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
	runStoreTests(t, testCases)
}

func TestStoreListPop(t *testing.T) {
	testCases := []storeTestCase{
		{
			name:  "pop operations on a non-list",
			setup: setupString("foo", "hello", time.Time{}),
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
			name:  "lpop on existing list",
			setup: setupList("foo", "a", "b", "c"),
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
			name:  "lpopn count less than length",
			setup: setupList("foo", "a", "b", "c", "d"),
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
			name:  "lpopn count greater than length",
			setup: setupList("foo", "a", "b", "c", "d"),
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
			name:  "lpopn zero count",
			setup: setupList("foo", "a", "b", "c"),
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
	runStoreTests(t, testCases)
}

func TestStoreListBLPop(t *testing.T) {
	testCases := []storeTestCase{
		{
			name: "blpop immediate pop",
			test: func(t *testing.T, s *store.Store) {
				_, err := s.RPush("foo", "a", "b")
				if err != nil {
					t.Fatalf("RPush() error = %v, want nil", err)
				}

				value, ok, err := s.BLPop("foo", 0)
				if err != nil {
					t.Fatalf("BLPop() error = %v, want nil", err)
				}

				if !ok {
					t.Fatalf("BLPop() ok = false, want true")
				}

				if value != "a" {
					t.Fatalf("BLPop() = %q, want %q", value, "a")
				}

				items, err := s.LRange("foo", 0, -1)
				if err != nil {
					t.Fatalf("LRange() error = %v, want nil", err)
				}

				want := []string{"b"}

				if !reflect.DeepEqual(items, want) {
					t.Fatalf("LRange() = %#v, want %#v", items, want)
				}
			},
		},
		{
			name: "blpop waits then rpush wakes it",
			test: func(t *testing.T, s *store.Store) {
				resultCh := make(chan blpopResult, 1)
				go func() {
					value, ok, err := s.BLPop("foo", 0)
					resultCh <- blpopResult{value, ok, err}
				}()
				time.Sleep(5 * time.Millisecond)

				n, err := s.RPush("foo", "a")
				if err != nil {
					t.Fatalf("RPush() error = %v, want nil", err)
				}

				if n != 1 {
					t.Fatalf("RPush() = %d, want 1", n)
				}

				//Len should be zero since the item was delivered directly to the waiter.
				ln, err := s.LLen("foo")
				if err != nil {
					t.Fatalf("LLen() error = %v, want nil", err)
				}

				if ln != 0 {
					t.Fatalf("LLen() length = %d, want 0", ln)
				}

				got := waitBLPop(t, resultCh)
				if got.err != nil {
					t.Fatalf("BLPop() error = %v, want nil", got.err)
				}

				if !got.ok {
					t.Fatalf("BLPop() ok = false, want true")
				}

				if got.value != "a" {
					t.Fatalf("BLPop() = %q, want %q", got.value, "a")
				}
			},
		},
		{
			name: "blpop timeout",
			test: func(t *testing.T, s *store.Store) {
				timeout := 3 * time.Millisecond
				start := time.Now()
				value, ok, blpopErr := s.BLPop("foo", timeout)
				elapsed := time.Since(start)

				n, err := s.RPush("foo", "a", "b")
				if err != nil {
					t.Fatalf("RPush() error = %v, want nil", err)
				}
				if n != 2 {
					t.Fatalf("RPush() = %d, want 2", n)
				}

				if blpopErr != nil {
					t.Fatalf("BLPop() error = %v, want nil", blpopErr)
				}

				if ok {
					t.Fatalf("BLPop() ok = true, want false")
				}

				if value != "" {
					t.Fatalf("BLPop() = %q, want %q", value, "")
				}

				if elapsed < timeout {
					t.Fatalf("BLPop() returned after %v, want at least %v", elapsed, timeout)
				}

				// Len should be two because the timed-out waiter should not receive pushed items.
				ln, err := s.LLen("foo")
				if err != nil {
					t.Fatalf("LLen() error = %v, want nil", err)
				}

				if ln != 2 {
					t.Fatalf("LLen() length = %d, want 2", ln)
				}
			},
		},
		{
			name: "blpop returns fifo waiter",
			test: func(t *testing.T, s *store.Store) {
				firstCh := make(chan blpopResult, 1)
				secondCh := make(chan blpopResult, 1)

				go func() {
					value, ok, err := s.BLPop("foo", 0)
					firstCh <- blpopResult{value, ok, err}
				}()
				time.Sleep(2 * time.Millisecond)

				go func() {
					value, ok, err := s.BLPop("foo", 0)
					secondCh <- blpopResult{value, ok, err}
				}()
				time.Sleep(2 * time.Millisecond)

				n, err := s.RPush("foo", "a")
				if err != nil {
					t.Fatalf("RPush() error = %v, want nil", err)
				}

				if n != 1 {
					t.Fatalf("RPush() = %d, want 1", n)
				}

				got := waitBLPop(t, firstCh)
				if got.err != nil {
					t.Fatalf("BLPop() error = %v, want nil", got.err)
				}

				if !got.ok {
					t.Fatalf("BLPop() ok = false, want true")
				}

				if got.value != "a" {
					t.Fatalf("BLPop() = %q, want %q", got.value, "a")
				}

				assertNoBLPop(t, secondCh)

				n, err = s.RPush("foo", "b")
				if err != nil {
					t.Fatalf("RPush() error = %v, want nil", err)
				}

				if n != 1 {
					t.Fatalf("RPush() = %d, want 1", n)
				}

				got = waitBLPop(t, secondCh)
				if got.err != nil {
					t.Fatalf("BLPop() error = %v, want nil", got.err)
				}

				if !got.ok {
					t.Fatalf("BLPop() ok = false, want true")
				}

				if got.value != "b" {
					t.Fatalf("BLPop() = %q, want %q", got.value, "b")
				}

				ln, err := s.LLen("foo")
				if err != nil {
					t.Fatalf("LLen() error = %v, want nil", err)
				}

				if ln != 0 {
					t.Fatalf("LLen() = %d, want 0", ln)
				}
			},
		},
	}

	runStoreTests(t, testCases)
}

func TestStoreReadAndPushExpired(t *testing.T) {
	expiredAt := time.Now().Add(-time.Second)

	testCases := []storeTestCase{
		{
			name:  "get string treats expired key as missing",
			setup: setupString("foo", "hello", expiredAt),
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
			name:  "rpush treats expired string as missing",
			setup: setupString("foo", "hello", expiredAt),
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
			name:  "lpush treats expired string as missing",
			setup: setupString("foo", "hello", expiredAt),
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
	runStoreTests(t, testCases)
}

func runStoreTests(t *testing.T, testCases []storeTestCase) {
	t.Helper()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := store.New()

			if tc.setup != nil {
				if err := tc.setup(s); err != nil {
					t.Fatalf("setup() error = %v, want nil", err)
				}
			}

			tc.test(t, s)
		})
	}
}

func waitBLPop(t *testing.T, ch <-chan blpopResult) blpopResult {
	t.Helper()

	select {
	case got := <-ch:
		return got
	case <-time.After(100 * time.Millisecond):
		t.Fatal("BLPop() did not return")
		return blpopResult{}
	}
}

func assertNoBLPop(t *testing.T, ch <-chan blpopResult) {
	t.Helper()

	select {
	case got := <-ch:
		t.Fatalf("BLPop() returned %#v, want still blocked", got)
	case <-time.After(10 * time.Millisecond):
	}
}

func setupString(key, value string, expiresAt time.Time) func(*store.Store) error {
	return func(s *store.Store) error {
		s.SetString(key, value, expiresAt)
		return nil
	}
}

func setupList(key string, items ...string) func(*store.Store) error {
	return func(s *store.Store) error {
		_, err := s.RPush(key, items...)
		return err
	}
}
