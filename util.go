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

func VarIntFromIOReader(r io.Reader) (int64, error) {
	b := &byter{r}
	return binary.ReadVarint(b)
}

// ReadVarUint() returns an uint64 by reading from r
func ReadVarUint(r io.Reader) (uint64, error) {

	buf := make([]byte, binary.MaxVarintLen64)
	counter := 0

	// read one byte at a a time
	for {
		b := make([]byte, 1)
		_, err := io.ReadAtLeast(r, b, 1)
		if err != nil {
			return 0, err
		}
		buf[counter] = b[0]

		// keep reading out bytes until no more data
		if b[0]&128 != 128 {
			break
		}

		counter++
	}

	decodedUInt, n := binary.Uvarint(buf)
	if n != counter+1 {
		return 0, fmt.Errorf("uvarint did not consume expected number of bytes")
	}

	return decodedUInt, nil

}

// WriteVarUint() writes a 64 bit unsigned integer to w
func WriteVarUint(w io.Writer, v uint64) error {

	buf := make([]byte, binary.MaxVarintLen64)
	varIntBytes := binary.PutUvarint(buf, v)
	n, err := w.Write(buf[0:varIntBytes])
	if err == nil && n != varIntBytes {
		err = errors.New("unexpected number of bytes written")
	}

	return err
}
