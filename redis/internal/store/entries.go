package store

import "time"

type Kind int

const (
	KindUnknown Kind = iota
	KindString
	KindList
)

type Entry struct {
	Value     Value
	ExpiresAt time.Time
}

type Value struct {
	Kind Kind
	Data any
}
