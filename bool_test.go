package schemer

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// base case, just make sure we can encode a boolean value, and subsequently read it back out
func TestDecodeBool1(t *testing.T) {

	bSchema := BoolSchema{WeakDecoding: false}

	var buf bytes.Buffer
	var err error
	var valueToEncode bool = false
	buf.Reset()

	fmt.Println("bool to bool")

	bSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	r := bytes.NewReader(buf.Bytes())

	var decodedValue1 bool
	err = bSchema.Decode(r, &decodedValue1)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != decodedValue1 {
		t.Errorf("Expected value")
	}

}

// make sure we can decode a boolean to the integer types
func TestDecodeBool2(t *testing.T) {

	bSchema := BoolSchema{WeakDecoding: true}

	var buf bytes.Buffer
	var err error
	var valueToEncode bool = true
	buf.Reset()

	bSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("bool to int")
	r := bytes.NewReader(buf.Bytes())

	var decodedValue1 int
	err = bSchema.Decode(r, &decodedValue1)
	if err != nil {
		t.Error(err)
	}

	if decodedValue1 != 1 {
		t.Errorf("Expected value")
	}

	fmt.Println("bool to uint")
	r = bytes.NewReader(buf.Bytes())

	var decodedValue2 uint
	err = bSchema.Decode(r, &decodedValue2)
	if err != nil {
		t.Error(err)
	}

	if decodedValue2 != 1 {
		t.Errorf("Expected value")
	}

}

// make sure we can decode a boolean value to a string, as long as weak decoding is on
func TestDecodeBool3(t *testing.T) {

	bSchema := BoolSchema{WeakDecoding: true}

	var buf bytes.Buffer
	var err error
	var valueToEncode bool = true
	buf.Reset()

	bSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("bool to string")
	r := bytes.NewReader(buf.Bytes())

	var decodedValue1 string
	err = bSchema.Decode(r, &decodedValue1)
	if err != nil {
		t.Error(err)
	}

	if strings.ToUpper(decodedValue1) != "TRUE" {
		t.Errorf("Expected value")
	}

}

// corner case:
// make sure library refuses to decode to an int when weak decoding is not enabled
func TestDecodeBool4(t *testing.T) {

	bSchema := BoolSchema{WeakDecoding: false}

	var buf bytes.Buffer
	var err error
	var valueToEncode bool = true
	buf.Reset()

	bSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("bool to int (w/o weak decoding enabled)")
	r := bytes.NewReader(buf.Bytes())

	var decodedValue1 int
	err = bSchema.Decode(r, &decodedValue1)
	if err == nil {
		t.Error("schemer library decoding failure; decoding bool to int should not be allowed w/o weak decoding")
	}

}

// corner case:
// make sure library refuses to decode to string when weak decoding is not enabled
func TestDecodeBool5(t *testing.T) {

	bSchema := BoolSchema{WeakDecoding: false}

	var buf bytes.Buffer
	var err error
	var valueToEncode bool = true
	buf.Reset()

	bSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("bool to string (w/o weak decoding enabled)")
	r := bytes.NewReader(buf.Bytes())

	var decodedValue1 string
	err = bSchema.Decode(r, &decodedValue1)
	if err == nil {
		t.Error("schemer library decoding failure; decoding bool to string should not be allowed w/o weak decoding")
	}

}
