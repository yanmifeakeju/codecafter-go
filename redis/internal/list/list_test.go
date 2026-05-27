package list_test

import (
	"testing"

	"github.com/yanmifeakeju/codecafter-go/redis/internal/list"
)

func TestList(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "push left",
			test: func(t *testing.T) {
				l := list.New("40", "50")
				if c := l.LPush("50"); c != 3 {
					t.Fatalf("LPush() expected count to be got=%d, expected=%d", c, 3)
				}

				items := l.Range(0, -1)

				if items[0] != "50" {
					t.Fatalf("LPush() got=%q, expected=%q", items[0], "50")
				}
			},
		},

		{
			name: "push right",
			test: func(t *testing.T) {
				l := list.New("40", "50")
				if c := l.RPush("50"); c != 3 {
					t.Fatalf("RPush() expected count to be got=%d, expected=%d", c, 3)
				}

				items := l.Range(0, -1)

				if items[l.Len()-1] != "50" {
					t.Fatalf("RPush() got=%q, expected=%q", items[l.Len()-1], "50")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}
