package schemer

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// byter wraps an io.Reader and provides a ReadByte() function
// Caution: ReadByte() usually implies that reading a single byte is fast, so
// please ensure that the underlying Reader can read individual bytes in
// sequence without major performance issues.
type byter struct {
	io.Reader
}

func (r byter) ReadByte() (byte, error) {
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

func (r byter) ReadVarint() (int64, error) {
	return binary.ReadVarint(r)
}

// ReadUvarint reads an Uvarint from r one byte at a time
func ReadUvarint(r io.Reader) (uint64, error) {

	rb := byter{r}
	buf := make([]byte, binary.MaxVarintLen64)

	// Read first byte into `buf`
	b, err := rb.ReadByte()
	if err != nil {
		return 0, err
	}
	buf[0] = b

	// Read subsequent bytes into `buf`
	i := 1
	for ; b&0x80 > 0; i++ {
		b, err = rb.ReadByte()
		if err != nil {
			return 0, err
		}
		buf[i] = b
	}

	decodedUInt, n := binary.Uvarint(buf)
	if n != i {
		return 0, fmt.Errorf("uvarint did not consume expected number of bytes")
	}

	return decodedUInt, nil

}

// WriteUvarint writes v to w as an Uvarint
func WriteUvarint(w io.Writer, v uint64) error {

	buf := make([]byte, binary.MaxVarintLen64)
	varIntBytes := binary.PutUvarint(buf, v)
	n, err := w.Write(buf[0:varIntBytes])
	if err == nil && n != varIntBytes {
		err = errors.New("unexpected number of bytes written")
	}

	return err
}
