package resp

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

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
	var err error

	switch v.Type {
	case SimpleString:
		nn, err := writeString(w, '+', v.Str)
		return int64(nn), err
	case SimpleError:
		nn, err := writeString(w, '-', v.Str)
		return int64(nn), err
	case Integers:
		nn, err := writeInt(w, ':', v.Num)
		return int64(nn), err
	case BulkString:
		if v.Null {
			nn, err := io.WriteString(w, "$-1\r\n")
			return int64(nn), err
		}

		n1, err := writeInt(w, '$', int64(len(v.Str)))
		if err != nil {
			return int64(n1), err
		}

		n2, err := io.WriteString(w, v.Str)
		if err != nil {
			return int64(n1 + n2), err
		}

		n3, err := w.Write([]byte{'\r', '\n'})
		return int64(n1 + n2 + n3), err
	case Array:
		if v.Null {
			nn, err := io.WriteString(w, "*-1\r\n")
			return int64(nn), err
		}

		n1, err := writeInt(w, '*', int64(len(v.Array)))
		if err != nil {
			return int64(n1), err
		}

		total := int64(n1)
		for _, elem := range v.Array {
			nn, elemErr := elem.WriteTo(w)
			total += nn
			if elemErr != nil {
				return total, elemErr
			}
		}
		return total, nil
	}

	return int64(0), err
}

func writeInt(w io.Writer, prefix byte, val int64) (int, error) {
	var buf [32]byte
	b := buf[:0]
	b = append(b, prefix)
	b = strconv.AppendInt(b, val, 10)
	b = append(b, '\r', '\n')
	return w.Write(b)
}

func writeString(w io.Writer, prefix byte, s string) (int, error) {
	var n int
	n1, err := w.Write([]byte{prefix})
	n += n1
	if err != nil {
		return n, err
	}

	n2, err := io.WriteString(w, s)
	n += n2
	if err != nil {
		return n, err
	}

	n3, err := w.Write([]byte{'\r', '\n'})
	n += n3
	return n, err
}
