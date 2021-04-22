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

// TestFixedIntegerSchema tests Schemer's FixedIntegerSchema function, making sure that it creates schemas as expected
func TestFixedIntegerSchema(t *testing.T) {

	var bs BasicSchema

	//----------------------------------------------------------
	// signed values
	//----------------------------------------------------------

	bs = FixedIntegerSchema(true, 8).(BasicSchema)
	if bs.Header != 1 {
		t.Errorf("invalid schema for signed 8 bit value")
	}

	bs = FixedIntegerSchema(true, 16).(BasicSchema)
	if bs.Header != 3 {
		t.Errorf("invalid schema for signed 16 bit value")
	}

	bs = FixedIntegerSchema(true, 32).(BasicSchema)
	if bs.Header != 5 {
		t.Errorf("invalid schema for signed 32 bit value")
	}

	bs = FixedIntegerSchema(true, 64).(BasicSchema)
	if bs.Header != 7 {
		t.Errorf("invalid schema for signed 64 bit value")
	}

	//----------------------------------------------------------
	// unsigned values
	//----------------------------------------------------------

	bs = FixedIntegerSchema(false, 8).(BasicSchema)
	if bs.Header != 0 {
		t.Errorf("invalid schema for unsigned 8 bit value")
	}

	bs = FixedIntegerSchema(false, 16).(BasicSchema)
	if bs.Header != 2 {
		t.Errorf("invalid schema for unsigned 16 bit value")
	}

	bs = FixedIntegerSchema(false, 32).(BasicSchema)
	if bs.Header != 4 {
		t.Errorf("invalid schema for unsigned 32 bit value")
	}

	bs = FixedIntegerSchema(false, 64).(BasicSchema)
	if bs.Header != 6 {
		t.Errorf("invalid schema for unsigned 64 bit value")
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

	var valueToEncode1 int64 = 1000
	buf.Reset()

	bs = FixedIntegerSchema(true, 64).(BasicSchema)
	bytesWritten, err := bs.Encode(&buf, valueToEncode1)
	if err != nil {
		t.Errorf("could not encode signed int 64 %v", err)
	}

	dumpBuffer(bytesWritten, buf)

	r := bytes.NewReader(buf.Bytes())

	fmt.Println("Testing decoding for go type int")
	var decodedValue1 int64
	err = bs.Decode(r, &decodedValue1)
	if err != nil {
		t.Errorf("could not decode signed int 64: %v", err)
	}

	if valueToEncode1 != decodedValue1 {
		t.Errorf("unexpected decoded value for 64 bit signed value. Expected value: %d; Decoded value: %d", valueToEncode1, decodedValue1)
	}

}
