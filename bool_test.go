package schemer

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// base case, just make sure we can encode a boolean value, and subsequently read it back out
// from an empty interface
func TestDecodeBool1(t *testing.T) {

	bSchema := BoolSchema{SchemaOptions{nullable: false}}

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

	var decodedValue1 interface{}
	err = bSchema.Decode(r, &decodedValue1)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != reflect.ValueOf(decodedValue1).Elem().Bool() {
		t.Errorf("Expected value")
	}

}

// make sure we can decode a boolean to the integer types
func TestDecodeBool2(t *testing.T) {

	bSchema := BoolSchema{SchemaOptions{weakDecoding: true}}

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

	bSchema := BoolSchema{SchemaOptions{weakDecoding: true}}

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

	bSchema := BoolSchema{SchemaOptions{weakDecoding: false}}

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

	bSchema := BoolSchema{SchemaOptions{weakDecoding: false}}

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
	schema := BoolSchema{SchemaOptions{nullable: false}}

	// encode it
	b, err := schema.MarshalSchemer()
	if err != nil {
		t.Error("cannot marshall schemer")
	}

	// make sure we can successfully decode it
	tmp, err := DecodeSchema(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err, "; cannot decode binary encoded bool")
	}

	decodedBoolSchema := tmp.(*BoolSchema)
	if decodedBoolSchema.Nullable() != schema.Nullable() {
		t.Error("unexpected value for BoolSchema")
	}

}

func TestDecodeBool7(t *testing.T) {

	boolSchema := BoolSchema{SchemaOptions{nullable: true}}

	fmt.Println("decode nil bool")

	var buf bytes.Buffer
	var err error
	var boolPtr *bool
	buf.Reset()

	err = boolSchema.Encode(&buf, boolPtr) // pass in nil pointer
	if err != nil {
		t.Error(err)
	}

	//------------

	r := bytes.NewReader(buf.Bytes())

	var boolToDecode bool
	var boolPtr2 *bool = &boolToDecode

	err = boolSchema.Decode(r, &boolPtr2)
	if err != nil {
		t.Error(err)
	}

	if boolPtr2 != nil {
		t.Error("unexpected value decoding null boolean")
	}

}

// make sure we can decode a nil bool into an interface
func TestDecodeBool7A(t *testing.T) {

	boolSchema := BoolSchema{SchemaOptions{
		nullable:     true,
		weakDecoding: false,
	}}

	fmt.Println("decode nil bool")

	var buf bytes.Buffer
	var err error
	var boolPtr *bool
	buf.Reset()

	err = boolSchema.Encode(&buf, boolPtr) // pass in nil pointer
	if err != nil {
		t.Error(err)
		return
	}

	// encode it
	b, err := boolSchema.MarshalSchemer()
	if err != nil {
		t.Error(err)
		return
	}

	//------------

	// make sure we can successfully decode it
	decodedSchema, err := DecodeSchema(bytes.NewReader(b))
	if err != nil {
		t.Error(err)
		return
	}

	r := bytes.NewReader(buf.Bytes())

	var decodedValue interface{} = "A" // just some random value

	err = decodedSchema.Decode(r, &decodedValue)
	if err != nil {
		t.Error(err)
	}

	if decodedValue != nil {
		t.Error("unexpected value decoding null boolean")
	}

}
