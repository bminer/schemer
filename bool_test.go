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

// TestDecodeBool6 makes sure we can encode and decode boolean schemas
func TestDecodeBool6(t *testing.T) {

	// setup an example schema
	schema := BoolSchema{IsNullable: false}

	// encode it
	b := schema.Bytes()

	// make sure we can successfully decode it
	tmp, err := NewSchema(b)
	decodedBoolSchema := tmp.(*BoolSchema)
	if err != nil {
		t.Error("cannot decode binary encoded bool")
	}
	if decodedBoolSchema.IsNullable != schema.IsNullable {
		t.Error("unexpected value for BoolSchema")
	}

}

func TestDecodeBool7(t *testing.T) {

	floatingPointSchema := BoolSchema{IsNullable: true}

	fmt.Println("decode nil bool")

	var buf bytes.Buffer
	var err error
	var boolPtr *bool
	buf.Reset()

	err = floatingPointSchema.Encode(&buf, boolPtr) // pass in nil pointer
	if err != nil {
		t.Error(err)
	}

	//------------

	r := bytes.NewReader(buf.Bytes())

	var boolToDecode bool
	var boolPtr2 *bool = &boolToDecode

	err = floatingPointSchema.Decode(r, &boolPtr2)
	if err != nil {
		t.Error(err)
	}

	if boolPtr2 != nil {
		t.Error("unexpected value decoding null boolean")
	}

}

func TestDecodeBool8(t *testing.T) {

	boolSchema := BoolSchema{IsNullable: true}

	b, _ := boolSchema.MarshalJSON()

	fmt.Print(string(b))

}
