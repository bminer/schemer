package schemer

import (
	"bytes"
	"fmt"
	"testing"
)

type Weekday int

const (
	Sunday    Weekday = 0
	Monday    Weekday = 1
	Tuesday   Weekday = 2
	Wednesday Weekday = 3
	Thursday  Weekday = 4
	Friday    Weekday = 5
	Saturday  Weekday = 6
)

// make sure we can encode/decode binary schemas for EnumSchema
func TestDecodeEnum1(t *testing.T) {

	// setup an example schema
	enumSchema := EnumSchema{IsNullable: false}

	// encode it
	b := enumSchema.Bytes()

	// make sure we can successfully decode it
	var decodedIntSchema EnumSchema
	var err error

	tmp, err := NewSchema(b)
	if err != nil {
		t.Error("cannot encode binary encoded enumSchema")
	}

	decodedIntSchema = tmp.(EnumSchema)

	// and then check the actual contents of the decoded schema
	// to make sure it contains the correct values
	if decodedIntSchema.IsNullable != enumSchema.IsNullable {
		t.Error("unexpected values when decoding binary EnumSchema")
	}

}

// TestDecodeEnum2 just tests the base case of decoding an enum to another enum
func TestDecodeEnum2(t *testing.T) {

	enumSchema := EnumSchema{IsNullable: false, WeakDecoding: false}

	// we have to manually fill in the writer's schema
	enumSchema.Values = make(map[int]string)
	enumSchema.Values[int(Sunday)] = "Sunday"
	enumSchema.Values[int(Monday)] = "Monday"
	enumSchema.Values[int(Tuesday)] = "Tuesday"
	enumSchema.Values[int(Wednesday)] = "Wednesday"
	enumSchema.Values[int(Thursday)] = "Thursday"
	enumSchema.Values[int(Friday)] = "Friday"
	enumSchema.Values[int(Saturday)] = "Saturday"

	var buf bytes.Buffer
	var err error
	var valueToEncode Weekday = 100 //Saturday
	buf.Reset()

	fmt.Println("Testing decoding enum value")

	enumSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	r := bytes.NewReader(buf.Bytes())

	var decodedValue1 Weekday
	err = enumSchema.Decode(r, &decodedValue1)
	if err != nil {
		t.Error(err)
	}

	if valueToEncode != decodedValue1 {
		t.Error("unexpected value during enum to enum decode")
	}

}

func TestDecodeEnum3(t *testing.T) {

	fmt.Println("decode nil enum")

	enumSchema := EnumSchema{IsNullable: true, WeakDecoding: false}

	var buf bytes.Buffer
	var err error
	var intPtr *int
	buf.Reset()

	//intPtr = nil
	err = enumSchema.Encode(&buf, intPtr)
	if err != nil {
		t.Error(err)
	}

	//------------

	r := bytes.NewReader(buf.Bytes())

	var intToDecodeTo int
	var intPtr2 *int = &intToDecodeTo

	err = enumSchema.Decode(r, &intPtr2)
	if err != nil {
		t.Error(err)
	}

	// floatPtr should be a nil pointer once we decoded it
	if intPtr2 != nil {
		t.Error("unexpected value decoding null enum")
	}

}
