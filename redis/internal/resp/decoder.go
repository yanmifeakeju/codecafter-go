package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
)

type Decoder struct {
	r *bufio.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: bufio.NewReader(r)}
}

func (dec *Decoder) Decode(val *Value) error {
	if val == nil {
		return errors.New("resp: Decode called with nil Value")
	}

	v, err := dec.decode()
	if err != nil {
		return err
	}

	*val = v

	return nil
}

func (dec *Decoder) decode() (Value, error) {
	b, err := dec.r.ReadByte()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return Value{}, io.EOF
		}
		return Value{}, fmt.Errorf("resp: read type byte: %w", err)
	}

	switch resType(b) {
	case BulkString:
		return readBulkStringValue(dec)
	case Array:
		return readArrayValue(dec)
	default:
		return Value{}, fmt.Errorf("resp: unsupported type byte %q", b)
	}
}

func (dec *Decoder) readValue(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}

	if _, err := io.ReadFull(dec.r, buf); err != nil {
		return fmt.Errorf("resp: read %d bytes: %w", len(buf), err)
	}

	return nil
}

func (dec *Decoder) readUntilCRLF() ([]byte, error) {
	b, err := dec.r.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("resp: read line: %w", err)
	}

	if len(b) < 2 || b[len(b)-2] != '\r' || b[len(b)-1] != '\n' {
		return nil, fmt.Errorf("resp: expected CRLF line ending")
	}

	return b[:len(b)-2], nil
}

func (dec *Decoder) readLength(kind string) (int64, error) {
	buf, err := dec.readUntilCRLF()
	if err != nil {
		return 0, err
	}

	n, err := parseLength(buf)
	if err != nil {
		return 0, fmt.Errorf("resp: invalid %s length %q: %w", kind, buf, err)
	}

	return n, nil
}

func parseLength(b []byte) (int64, error) {
	if len(b) == 0 {
		return 0, errors.New("empty length")
	}

	negative := b[0] == '-'
	if negative {
		if len(b) == 1 {
			return 0, errors.New("missing length digits")
		}
		b = b[1:]
	}

	// Parse the magnitude as uint64 so we can represent the absolute value of
	// math.MinInt64, which is one greater than math.MaxInt64.
	var n uint64
	max := uint64(math.MaxInt64)
	if negative {
		max++
	}

	for _, c := range b {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid digit %q", c)
		}

		digit := uint64(c - '0')
		if n > (max-digit)/10 {
			return 0, errors.New("length out of range")
		}
		n = n*10 + digit
	}

	if negative {
		if n == max {
			return -int64(max-1) - 1, nil
		}
		return -int64(n), nil
	}

	return int64(n), nil
}

func readBulkStringValue(dec *Decoder) (Value, error) {
	n, err := dec.readLength("bulk string")
	if err != nil {
		return Value{}, err
	}

	if n == -1 {
		return Value{Type: BulkString, Null: true}, nil
	}
	if n < -1 {
		return Value{}, fmt.Errorf("resp: invalid bulk string length %d", n)
	}

	if n > int64(math.MaxInt)-2 {
		return Value{}, fmt.Errorf("resp: bulk string length %d is too large", n)
	}

	buf := make([]byte, int(n))
	if err := dec.readValue(buf); err != nil {
		return Value{}, err
	}

	var crlf [2]byte
	if err := dec.readValue(crlf[:]); err != nil {
		return Value{}, err
	}

	if crlf[0] != '\r' || crlf[1] != '\n' {
		return Value{}, fmt.Errorf("resp: expected CRLF after bulk string")
	}

	return BulkStringValue(string(buf[:n])), nil
}

func readArrayValue(dec *Decoder) (Value, error) {
	n, err := dec.readLength("array")
	if err != nil {
		return Value{}, err
	}

	if n == -1 {
		return Value{Type: Array, Null: true}, nil
	}

	if n < -1 {
		return Value{}, fmt.Errorf("resp: invalid array length %d", n)
	}

	if n > int64(math.MaxInt) {
		return Value{}, fmt.Errorf("resp: array length %d is too large", n)
	}

	items := make([]Value, 0, int(n))
	for i := range n {
		item, err := dec.decode()
		if err != nil {
			return Value{}, fmt.Errorf("resp: decode array element %d: %w", i, err)
		}
		items = append(items, item)
	}

	return ArrayValue(items), nil
}
