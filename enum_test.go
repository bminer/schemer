package schemer

import (
	"bytes"
	"fmt"
	"strconv"
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
	enumSchema := EnumSchema{SchemaOptions: SchemaOptions{Nullable: false}}

	// encode it
	b := enumSchema.MarshalSchemer()

	// make sure we can successfully decode it

	tmp, err := DecodeSchema(b)
	if err != nil {
		t.Error("cannot encode binary encoded enumSchema")
	}

	decodedEnumSchema := tmp.(*EnumSchema)

	// and then check the actual contents of the decoded schema
	// to make sure it contains the correct values
	if decodedEnumSchema.SchemaOptions.Nullable != enumSchema.SchemaOptions.Nullable {
		t.Error("unexpected values when decoding binary EnumSchema")
	}

}

// TestDecodeEnum2 just tests the base case of decoding an enum to another enum
func TestDecodeEnum2(t *testing.T) {

	enumSchema := EnumSchema{SchemaOptions: SchemaOptions{Nullable: false, WeakDecoding: false}}

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
	var valueToEncode Weekday = Saturday
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

// TestDecodeEnum3 tests that we can encode and decode a nullable enum value
func TestDecodeEnum3(t *testing.T) {

	fmt.Println("decode nil enum")

	enumSchema := EnumSchema{SchemaOptions: SchemaOptions{Nullable: true, WeakDecoding: false}}

	var buf bytes.Buffer
	var err error
	var intPtr *int
	buf.Reset()

	intPtr = nil
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

// TestDecodeEnum4 tests that we can decode an enum to a string,
// when we have the map
func TestDecodeEnum4(t *testing.T) {

	fmt.Println("decode enum to string")

	enumSchema := EnumSchema{SchemaOptions: SchemaOptions{Nullable: true, WeakDecoding: true}}

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
	var valueToEncode Weekday = Tuesday
	buf.Reset()

	err = enumSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	//------------

	r := bytes.NewReader(buf.Bytes())

	var enumToDecodeTo string

	err = enumSchema.Decode(r, &enumToDecodeTo)
	if err != nil {
		t.Error(err)
	}

	// floatPtr should be a nil pointer once we decoded it
	if enumSchema.Values[int(valueToEncode)] != enumToDecodeTo {
		t.Error("unexpected value decoding enum to string")
	}

}

// TestDecodeEnum5 tests that we can decode an enum to a string,
// when no map is present
func TestDecodeEnum5(t *testing.T) {

	fmt.Println("decode enum to string")

	enumSchema := EnumSchema{SchemaOptions: SchemaOptions{Nullable: true, WeakDecoding: true}}

	var buf bytes.Buffer
	var err error
	var valueToEncode Weekday = Tuesday
	buf.Reset()

	err = enumSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	//------------

	r := bytes.NewReader(buf.Bytes())

	var enumToDecodeTo string

	err = enumSchema.Decode(r, &enumToDecodeTo)
	if err != nil {
		t.Error(err)
	}

	// floatPtr should be a nil pointer once we decoded it
	if strconv.Itoa(int(valueToEncode)) != enumToDecodeTo {
		t.Error("unexpected value decoding enum to string")
	}

}

// TestDecodeEnum6 makes sure that decoding an enumerated type that is not in the writer's
// schema will throw an errow
func TestDecodeEnum6(t *testing.T) {

	enumSchema := EnumSchema{SchemaOptions: SchemaOptions{Nullable: false, WeakDecoding: false}}

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
	var valueToEncode Weekday = 100 // intentionally encode an invalid value.. it should be caught on the decode...
	buf.Reset()

	fmt.Println("Testing decoding invalid enum value")

	enumSchema.Encode(&buf, valueToEncode)
	if err != nil {
		t.Error(err)
	}

	r := bytes.NewReader(buf.Bytes())

	var decodedValue1 Weekday
	err = enumSchema.Decode(r, &decodedValue1)
	if err == nil {
		t.Error("schemer failure; invalid enum allowed to be decoded")
	}

}

func TestEnumWriter(t *testing.T) {

	enumSchema := EnumSchema{SchemaOptions: SchemaOptions{Nullable: false, WeakDecoding: false}}

	// we have to manually fill in the writer's schema
	enumSchema.Values = make(map[int]string)
	enumSchema.Values[int(Sunday)] = "Sunday"
	enumSchema.Values[int(Monday)] = "Monday"
	enumSchema.Values[int(Tuesday)] = "Tuesday"
	enumSchema.Values[int(Wednesday)] = "Wednesday"
	enumSchema.Values[int(Thursday)] = "Thursday"
	enumSchema.Values[int(Friday)] = "Friday"
	enumSchema.Values[int(Saturday)] = "Saturday"

	enumToDecode := Saturday
	binaryReaderSchema := enumSchema.MarshalSchemer()

	var encodedData bytes.Buffer

	err := enumSchema.Encode(&encodedData, enumToDecode)
	if err != nil {
		t.Error(err)
	}

	saveToDisk("/tmp/Enum.schema", binaryReaderSchema)
	saveToDisk("/tmp/Enum.data", encodedData.Bytes())

}

func TestEnumReader(t *testing.T) {

	var enumToDecode int

	binarywriterSchema := readFromDisk("/tmp/Enum.schema")
	writerSchema, err := DecodeSchema(binarywriterSchema)
	if err != nil {
		t.Error("cannot create writerSchema from raw binary data")
	}

	encodedData := readFromDisk("/tmp/Enum.data")
	r := bytes.NewReader(encodedData)

	err = writerSchema.Decode(r, &enumToDecode)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(enumToDecode)

}
func TestEnumSerialize(t *testing.T) {
	TestEnumWriter(t)
	TestEnumReader(t)
}
