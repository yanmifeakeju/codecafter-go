package list

type List struct {
	items []string
}

func New(items ...string) *List {
	return &List{items: append([]string(nil), items...)}
}

func (l *List) LPush(items ...string) int {
	for _, item := range items {
		l.items = append([]string{item}, l.items...)
	}
	return len(l.items)
}

func (l *List) RPush(items ...string) int {
	l.items = append(l.items, items...)
	return len(l.items)
}

func (l *List) Range(start, stop int) []string {
	n := len(l.items)

	if n == 0 {
		return []string{}
	}

	// Translate negative indexes from the end.
	if start < 0 {
		start = n + start
	}
	if stop < 0 {
		stop = n + stop
	}

	if start < 0 {
		start = 0
	}
	if stop >= n {
		stop = n - 1
	}

	if start > stop || start >= n {
		return []string{}
	}

	return append([]string(nil), l.items[start:stop+1]...)
}

func (l *List) LPop() (string, bool) {
	if len(l.items) == 0 {
		return "", false
	}

	result := l.items[0]
	l.items = l.items[1:]
	return result, true
}

func (l *List) LPopN(count int) []string {
	if count <= 0 || len(l.items) == 0 {
		return []string{}
	}

	if count > len(l.items) {
		count = len(l.items)
	}

	items := append([]string(nil), l.items[:count]...)
	l.items = l.items[count:]

	return items
}

func (l *List) Len() int {
	return len(l.items)
}
