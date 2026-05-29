package store

import (
	"errors"
	"sync"
	"time"

	"github.com/yanmifeakeju/codecafter-go/redis/internal/list"
)

var ErrWrongType = errors.New("wrong type")

type Store struct {
	values  map[string]Entry
	waiters map[string][]*waiter
	mu      sync.Mutex
}

type waiter struct {
	ch chan string
}

func New() *Store {
	return &Store{
		values:  map[string]Entry{},
		waiters: map[string][]*waiter{},
	}
}

func (s *Store) SetString(key, value string, expiresAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.values[key] = Entry{Value: Value{Kind: KindString, Data: value}, ExpiresAt: expiresAt}
}

func (s *Store) GetString(key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.getEntry(key)
	if !ok {
		return "", false, nil
	}

	if entry.Value.Kind != KindString {
		return "", false, ErrWrongType
	}

	str, ok := entry.Value.Data.(string)
	if !ok {
		return "", false, ErrWrongType
	}

	return str, true, nil
}

func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.values[key]; !ok {
		return false
	}

	delete(s.values, key)
	return true
}

func (s *Store) RPush(key string, items ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pushed := len(items)
	entry, ok := s.getEntry(key)
	var l *list.List
	created := false

	if !ok {
		l = list.New()
		created = true
	} else {
		if entry.Value.Kind != KindList {
			return 0, ErrWrongType
		}

		l, ok = entry.Value.Data.(*list.List)
		if !ok {
			return 0, ErrWrongType
		}
	}

	for len(items) > 0 && len(s.waiters[key]) > 0 {
		waiter := s.waiters[key][0]
		s.waiters[key] = s.waiters[key][1:]

		waiter.ch <- items[0]
		items = items[1:]
	}

	if len(items) == 0 {
		return l.Len() + pushed, nil
	}

	delivered := pushed - len(items)
	n := l.RPush(items...)
	if created {
		s.values[key] = Entry{
			Value: Value{
				Kind: KindList,
				Data: l,
			},
		}
	}

	return n + delivered, nil
}

func (s *Store) LPush(key string, items ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pushed := len(items)
	entry, ok := s.getEntry(key)
	var l *list.List
	created := false

	if !ok {
		l = list.New()
		created = true
	} else {
		if entry.Value.Kind != KindList {
			return 0, ErrWrongType
		}

		l, ok = entry.Value.Data.(*list.List)
		if !ok {
			return 0, ErrWrongType
		}
	}

	for len(items) > 0 && len(s.waiters[key]) > 0 {
		waiter := s.waiters[key][0]
		s.waiters[key] = s.waiters[key][1:]

		waiter.ch <- items[len(items)-1]
		items = items[:len(items)-1]
	}

	if len(items) == 0 {
		return l.Len() + pushed, nil
	}

	delivered := pushed - len(items)
	n := l.LPush(items...)
	if created {
		s.values[key] = Entry{
			Value: Value{
				Kind: KindList,
				Data: l,
			},
		}
	}

	return n + delivered, nil
}

func (s *Store) LRange(key string, start, end int) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var res []string
	entry, ok := s.getEntry(key)

	if !ok {
		return res, nil
	}

	if entry.Value.Kind != KindList {
		return res, ErrWrongType
	}

	l, ok := entry.Value.Data.(*list.List)
	if !ok {
		return res, ErrWrongType
	}

	if l.Len() == 0 {
		return res, nil
	}

	res = l.Range(start, end)

	return res, nil
}

func (s *Store) LLen(key string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.getEntry(key)
	if !ok {
		return 0, nil
	}

	if entry.Value.Kind != KindList {
		return 0, ErrWrongType
	}

	l, ok := entry.Value.Data.(*list.List)
	if !ok {
		return 0, ErrWrongType
	}

	return l.Len(), nil
}

func (s *Store) LPop(key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.lpopLocked(key)
}

func (s *Store) LPopN(key string, count int) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.getEntry(key)
	if !ok {
		return []string{}, nil
	}

	if entry.Value.Kind != KindList {
		return nil, ErrWrongType
	}

	l, ok := entry.Value.Data.(*list.List)
	if !ok {
		return nil, ErrWrongType
	}

	return l.LPopN(count), nil
}

func (s *Store) getEntry(key string) (Entry, bool) {
	entry, ok := s.values[key]
	if !ok {
		return Entry{}, false
	}

	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		delete(s.values, key)
		return Entry{}, false
	}

	return entry, true
}

func (s *Store) BLPop(key string, timeout time.Duration) (string, bool, error) {
	s.mu.Lock()

	// First try immediate BLPop
	value, ok, err := s.lpopLocked(key)
	if err != nil {
		s.mu.Unlock()
		return "", false, err
	}

	if ok {
		s.mu.Unlock()
		return value, true, nil
	}

	waiter := &waiter{ch: make(chan string, 1)}
	s.waiters[key] = append(s.waiters[key], waiter)
	s.mu.Unlock()

	if timeout == 0 {
		value = <-waiter.ch
		return value, true, nil
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case value := <-waiter.ch:
		return value, true, nil
	case <-timer.C:
		if s.removeWaiter(key, waiter) {
			return "", false, nil
		}

		// It was already removed by <R/L>PUSH, so a value is on the channel.
		value := <-waiter.ch
		return value, true, nil
	}
}

func (s *Store) lpopLocked(key string) (string, bool, error) {
	entry, ok := s.getEntry(key)
	if !ok {
		return "", false, nil
	}

	if entry.Value.Kind != KindList {
		return "", false, ErrWrongType
	}

	l, ok := entry.Value.Data.(*list.List)
	if !ok {
		return "", false, ErrWrongType
	}

	item, ok := l.LPop()

	return item, ok, nil
}

func (s *Store) removeWaiter(key string, target *waiter) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	waiters := s.waiters[key]
	for i, w := range waiters {
		if w == target {
			s.waiters[key] = append(waiters[:i], waiters[i+1:]...)
			if len(s.waiters[key]) == 0 {
				delete(s.waiters, key)
			}
			return true
		}
	}

	return false
}
