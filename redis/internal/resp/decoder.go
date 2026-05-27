package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
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

	n, err := strconv.ParseInt(string(buf), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("resp: invalid %s length %q: %w", kind, string(buf), err)
	}

	return n, nil
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

	if n > int64(^uint(0)>>1)-2 {
		return Value{}, fmt.Errorf("resp: bulk string length %d is too large", n)
	}

	buf := make([]byte, int(n)+2)
	if err := dec.readValue(buf); err != nil {
		return Value{}, err
	}

	if buf[n] != '\r' || buf[n+1] != '\n' {
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

	if n > int64(^uint(0)>>1) {
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
