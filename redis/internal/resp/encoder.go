package resp

import (
	"bufio"
	"io"
)

type Encoder struct {
	w *bufio.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: bufio.NewWriter(w)}
}

func (enc *Encoder) Encode(v Value) error {
	if _, err := v.WriteTo(enc.w); err != nil {
		return err
	}
	return enc.w.Flush()
}
