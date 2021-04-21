package schemer

import (
	"bytes"
	"fmt"
	"testing"
)

func TestShowIntSize(t *testing.T) {
	if ArchitectureIs64Bits() {
		fmt.Println("Size of int in bits is 64")

	} else {
		fmt.Println("Size of int in bits is 32")

	}
}

// TestCreateSchemaForFixedLenInt tests schemer's ability to correctly generate
// a schema for a fixed length signed or unsigned integer
func TestCreateSchemaForFixedLenInt(t *testing.T) {

	var v1 int
	var v2 int8
	var v3 int16
	var v4 int32
	var v5 int64
	var v6 uint
	var v7 uint8
	var v8 uint16
	var v9 uint32
	var v10 uint64

	var bs BasicSchema
	var err error

	// don't forget ints can either be 32 or 64 bits wide, based on the architecture
	bs.Header, err = bs.FixedSizeInt(v1)
	if ArchitectureIs64Bits() {
		if bs.Header != 7 || err != nil {
			t.Errorf("invalid schema for go type: int")
		}
	} else {
		if bs.Header != 5 || err != nil {
			t.Errorf("invalid schema for go type: int")
		}
	}

	bs.Header, err = bs.FixedSizeInt(v2)
	if bs.Header != 1 || err != nil {
		t.Errorf("invalid schema for go type: int8")
	}

	bs.Header, err = bs.FixedSizeInt(v3)
	if bs.Header != 3 || err != nil {
		t.Errorf("invalid schema for go type: int16")
	}

	bs.Header, err = bs.FixedSizeInt(v4)
	if bs.Header != 5 || err != nil {
		t.Errorf("invalid schema for go type: int32")
	}

	bs.Header, err = bs.FixedSizeInt(v5)
	if bs.Header != 7 || err != nil {
		t.Errorf("invalid schema for go type: int64")
	}

	bs.Header, err = bs.FixedSizeInt(v6)
	if ArchitectureIs64Bits() {
		if bs.Header != 6 || err != nil {
			t.Errorf("invalid schema for go type: uint")
		}
	} else {
		if bs.Header != 4 || err != nil {
			t.Errorf("invalid schema for go type: uint")
		}
	}

	bs.Header, err = bs.FixedSizeInt(v7)
	if bs.Header != 0 || err != nil {
		t.Errorf("invalid schema for go type: uint8")
	}

	bs.Header, err = bs.FixedSizeInt(v8)
	if bs.Header != 2 || err != nil {
		t.Errorf("invalid schema for go type: uint16")
	}

	bs.Header, err = bs.FixedSizeInt(v9)
	if bs.Header != 4 || err != nil {
		t.Errorf("invalid schema for go type: uint32")
	}

	bs.Header, err = bs.FixedSizeInt(v10)
	if bs.Header != 6 || err != nil {
		t.Errorf("invalid schema for go type: uint64")
	}

}

func dumpBuffer(bytesWritten int, buf bytes.Buffer) {

	fmt.Printf("total bytes written: %d \n", bytesWritten)
	for i := 0; i < bytesWritten; i++ {
		fmt.Printf("byte%d: %d\n", i, buf.Bytes()[i])
	}

}

func TestEncodeFixedLenInt(t *testing.T) {

	var buf bytes.Buffer
	var bs BasicSchema
	var err error

	fmt.Println("---------------------------------------")
	fmt.Println("Testing encoding for go type int")

	var v1 int = 100

	bs.Header, err = bs.FixedSizeInt(v1)
	if err != nil {
		t.Errorf("cannot create schema for go type int")
	}

	bytesWritten, err := bs.Encode(&buf, v1)
	if err != nil {
		t.Errorf("could not encode int: %v", err)
	}

	dumpBuffer(bytesWritten, buf)

	fmt.Println("---------------------------------------")

	fmt.Println("Testing encoding for go type int8")

	var v2 int8 = 101

	bs.Header, err = bs.FixedSizeInt(v2)
	if err != nil {
		t.Errorf("cannot create schema for go type int8")
	}

	bytesWritten, err = bs.Encode(&buf, v2)
	if err != nil {
		t.Errorf("could not encode int: %v", err)
	}

	dumpBuffer(bytesWritten, buf)

	fmt.Println("---------------------------------------")

	fmt.Println("Testing encoding for go type int16")

	var v3 int16 = 102

	bs.Header, err = bs.FixedSizeInt(v3)
	if err != nil {
		t.Errorf("cannot create schema for go type int16")
	}

	bytesWritten, err = bs.Encode(&buf, v3)
	if err != nil {
		t.Errorf("could not encode int: %v", err)
	}

	dumpBuffer(bytesWritten, buf)

	fmt.Println("---------------------------------------")

}
