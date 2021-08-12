package schemer

import (
	"encoding/binary"
	"io"
)

type byter struct {
	io.Reader
}

func (r *byter) ReadByte() (byte, error) {
	var buf [1]byte
	n, err := r.Reader.Read(buf[:])
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, io.ErrNoProgress
	}
	return buf[0], nil
}

func VarIntFromIOReader(r io.Reader) (int64, error) {
	b := &byter{r}
	return binary.ReadVarint(b)
}
