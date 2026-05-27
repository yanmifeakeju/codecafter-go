package resp

import (
	"fmt"
	"io"
	"strings"
)

// import (
// 	"bufio"
// 	"errors"
// 	"fmt"
// 	"io"
// 	"strconv"
// 	"strings"
// )

type resType byte

const (
	SimpleString resType = '+'
	SimpleError  resType = '-'
	Integers     resType = ':'
	BulkString   resType = '$'
	Array        resType = '*'
)

type Value struct {
	Type  resType
	Str   string
	Num   int64
	Array []Value
	Null  bool
}

func (v Value) String() string {
	switch v.Type {
	case SimpleString:
		return fmt.Sprintf("SimpleString(%q)", v.Str)
	case SimpleError:
		return fmt.Sprintf("Error(%q)", v.Str)
	case Integers:
		return fmt.Sprintf("Integer(%d)", v.Num)
	case BulkString:
		if v.Null {
			return "BulkString(nil)"
		}
		return fmt.Sprintf("BulkString(%q)", v.Str)
	case Array:
		if v.Null {
			return "Array(nil)"
		}
		elems := make([]string, len(v.Array))
		for i, e := range v.Array {
			elems[i] = e.String()
		}
		return fmt.Sprintf("Array[%s]", strings.Join(elems, ", "))
	default:
		return "Unknown"
	}
}

func SimpleStringValue(s string) Value {
	return Value{Type: SimpleString, Str: s}
}

func ErrorValue(s string) Value {
	return Value{Type: SimpleError, Str: s}
}

func IntValue(n int64) Value {
	return Value{Type: Integers, Num: n}
}

func BulkStringValue(s string) Value {
	return Value{Type: BulkString, Str: s}
}

func NullBulkString() Value {
	return Value{Type: BulkString, Null: true}
}

func ArrayValue(v []Value) Value {
	return Value{Type: Array, Array: v}
}

func (v Value) WriteTo(w io.Writer) (int64, error) {
	var n int
	var err error

	switch v.Type {
	case SimpleString:
		n, err = fmt.Fprintf(w, "+%s\r\n", v.Str)
	case SimpleError:
		n, err = fmt.Fprintf(w, "-%s\r\n", v.Str)
	case Integers:
		n, err = fmt.Fprintf(w, ":%d\r\n", v.Num)
	case BulkString:
		if v.Null {
			n, err = fmt.Fprint(w, "$-1\r\n")
		} else {
			n, err = fmt.Fprintf(w, "$%d\r\n%s\r\n", len(v.Str), v.Str)
		}
	case Array:
		if v.Null {
			n, err = fmt.Fprint(w, "*-1\r\n")
		} else {
			n, err = fmt.Fprintf(w, "*%d\r\n", len(v.Array))
			if err != nil {
				return int64(n), err
			}
			for _, elem := range v.Array {
				nn, elemErr := elem.WriteTo(w)
				n += int(nn)
				if elemErr != nil {
					return int64(n), elemErr
				}
			}
		}
	}

	return int64(n), err
}
