package schemer

import (
	"bytes"
	"fmt"
	"testing"
)

// if true, we will dump out the bytes written during encoding to stdout
var verboseOutput bool = false

func TestShowIntSize(t *testing.T) {

	if !verboseOutput {
		return
	}

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

	if !verboseOutput {
		return
	}

	fmt.Printf("total bytes written: %d \n", bytesWritten)
	for i := 0; i < bytesWritten; i++ {
		fmt.Printf("byte%d: %d\n", i, buf.Bytes()[i])
	}

}

func TestEncodeFixedLenInt(t *testing.T) {

	var buf bytes.Buffer
	var bs BasicSchema
	var err error
	var jSON []byte

	fmt.Println("---------------------------------------")
	fmt.Println("Testing encoding for go type int64")

	var valueToEncode1 int64 = 1000
	buf.Reset()

	bs = FixedIntegerSchema(true, 64).(BasicSchema)
	bytesWritten, err := bs.Encode(&buf, valueToEncode1)
	if err != nil {
		t.Errorf("could not encode signed int 64 %v", err)
	}

	dumpBuffer(bytesWritten, buf)

	r := bytes.NewReader(buf.Bytes())

	fmt.Println("Testing decoding for go type int64")
	var decodedValue1 int64
	err = bs.Decode(r, &decodedValue1)
	if err != nil {
		t.Errorf("could not decode signed int64: %v", err)
	}

	if valueToEncode1 != decodedValue1 {
		t.Errorf("unexpected decoded value for 64 bit signed value. Expected value: %d; Decoded value: %d", valueToEncode1, decodedValue1)
	}

	// test generating JSON
	jSON, err = bs.MarshalJSON()
	if err != nil {
		t.Errorf("could not marshal JSON for signed int 64 %v", err)
	}
	fmt.Println("JSON: " + string(jSON))

	// TODO: test creating type from JSON

	fmt.Println("---------------------------------------")
	fmt.Println("Testing encoding for go type int32")

	var valueToEncode2 int32 = 4000
	buf.Reset()

	bs = FixedIntegerSchema(true, 32).(BasicSchema)
	bytesWritten, err = bs.Encode(&buf, valueToEncode2)
	if err != nil {
		t.Errorf("could not encode signed int 32 %v", err)
	}

	dumpBuffer(bytesWritten, buf)

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("Testing decoding for go type int32")
	var decodedValue2 int32
	err = bs.Decode(r, &decodedValue2)
	if err != nil {
		t.Errorf("could not decode signed int32: %v", err)
	}

	if valueToEncode2 != decodedValue2 {
		t.Errorf("unexpected decoded value for 32 bit signed value. Expected value: %d; Decoded value: %d", valueToEncode2, decodedValue2)
	}

	// test generating JSON
	jSON, err = bs.MarshalJSON()
	if err != nil {
		t.Errorf("could not marshal JSON for signed int 64 %v", err)
	}
	fmt.Println("JSON: " + string(jSON))

	// TODO: test creating type from JSON

	fmt.Println("---------------------------------------")
	fmt.Println("Testing encoding for go type int16")

	var valueToEncode3 int16 = 5000
	buf.Reset()

	bs = FixedIntegerSchema(true, 16).(BasicSchema)
	bytesWritten, err = bs.Encode(&buf, valueToEncode3)
	if err != nil {
		t.Errorf("could not encode signed int 16 %v", err)
	}

	dumpBuffer(bytesWritten, buf)

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("Testing decoding for go type int16")
	var decodedValue3 int16
	err = bs.Decode(r, &decodedValue3)
	if err != nil {
		t.Errorf("could not decode signed int16: %v", err)
	}

	if valueToEncode3 != decodedValue3 {
		t.Errorf("unexpected decoded value for 16 bit signed value. Expected value: %d; Decoded value: %d", valueToEncode3, decodedValue3)
	}

	// test generating JSON
	jSON, err = bs.MarshalJSON()
	if err != nil {
		t.Errorf("could not marshal JSON for signed int 64 %v", err)
	}
	fmt.Println("JSON: " + string(jSON))

	// TODO: test creating type from JSON

	fmt.Println("---------------------------------------")
	fmt.Println("Testing encoding for go type int8")

	var valueToEncode4 int8 = 101
	buf.Reset()

	bs = FixedIntegerSchema(true, 8).(BasicSchema)
	bytesWritten, err = bs.Encode(&buf, valueToEncode4)
	if err != nil {
		t.Errorf("could not encode signed int 8 %v", err)
	}

	dumpBuffer(bytesWritten, buf)

	r = bytes.NewReader(buf.Bytes())

	fmt.Println("Testing decoding for go type int8")
	var decodedValue4 int8
	err = bs.Decode(r, &decodedValue4)
	if err != nil {
		t.Errorf("could not decode signed int8: %v", err)
	}

	if valueToEncode4 != decodedValue4 {
		t.Errorf("unexpected decoded value for 8 bit signed value. Expected value: %d; Decoded value: %d", valueToEncode4, decodedValue4)
	}

	// test generating JSON
	jSON, err = bs.MarshalJSON()
	if err != nil {
		t.Errorf("could not marshal JSON for signed int 64 %v", err)
	}
	fmt.Println("JSON: " + string(jSON))

	// TODO: test creating type from JSON

}
