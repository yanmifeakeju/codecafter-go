package dict

import "time"

type Kind int

const (
	KindUnknown Kind = iota
	KindString
	KindList
)

type Dict struct {
	values map[string]Entry
}

type Entry struct {
	Value     Value
	ExpiresAt time.Time
}

type Value struct {
	Kind Kind
	Data any
}

func New() *Dict {
	return &Dict{
		values: map[string]Entry{},
	}
}

func (d *Dict) Set(key string, entry Entry) {
	d.values[key] = entry
}

func (d *Dict) Get(key string) (Entry, bool) {
	entry, ok := d.values[key]

	return entry, ok
}

func (d *Dict) Delete(key string) bool {
	if _, ok := d.values[key]; !ok {
		return false
	}

	delete(d.values, key)
	return true
}

func (k Kind) String() string {
	switch k {
	case KindString:
		return "string"
	case KindList:
		return "list"
	default:
		return "unknown"
	}
}
