package schemer

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestDecodeFixedObject1(t *testing.T) {

	type TestStruct struct {
		A string
	}

	var testStruct TestStruct

	// setup an example schema
	fixedObjectSchema := SchemaOf(testStruct)

	// encode it
	b := fixedObjectSchema.MarshalSchemer()

	tmp, err := DecodeSchema(b)
	if err != nil {
		t.Error("cannot encode binary encoded FixedObjectSchema")
	}

	decodedSchema := tmp.(*FixedObjectSchema)

	// and then check the actual contents of the decoded schema
	// to make sure it contains the correct values
	if decodedSchema.Nullable() != fixedObjectSchema.Nullable() {
		t.Error("unexpected values when decoding binary FixedObjectSchema")
	}

}

// TestDecodeFixedObject2 tests encoding a nil
func TestDecodeFixedObject2(t *testing.T) {

	type StructToEncode struct {
		XX *int `schemer:""`
	}

	var structToEncode *StructToEncode = &(StructToEncode{})

	var buf bytes.Buffer
	var err error

	buf.Reset()

	fixedObjectSchema := SchemaOf(&structToEncode)

	// test overririding the nullability of the string...
	//fixedObjectSchema.(*FixedObjectSchema).Fields[0].Schema.(*VarLenStringSchema).IsNullable = false

	err = fixedObjectSchema.Encode(&buf, structToEncode)
	if err != nil {
		t.Error(err)
	}

	r := bytes.NewReader(buf.Bytes())

	type StructToDecode struct {
		YY *int `schemer:"[XX]"`
	}

	var structToDecode = StructToDecode{}

	fmt.Println(structToEncode)
	json, _ := fixedObjectSchema.MarshalJSON()
	fmt.Println(string(json))
	fmt.Println(buf.Bytes())

	err = fixedObjectSchema.Decode(r, &structToDecode)
	if err != nil {
		t.Error(err)
	}

	if structToDecode.YY == nil {
		fmt.Println("YY decoded as null value...")
	} else {
		fmt.Println(*structToDecode.YY)
	}

}

// TestDecodeFixedObject5 tests our ability to decode objects to other objects, using struct tags....
func TestDecodeFixedObject5(t *testing.T) {

	type SourceStruct struct {
		FName     string `schemer:"[A,B,C,FirstName]"`
		LName     string `schemer:"[D,E,F,LastName]"`
		AgeInLife int    `schemer:"Age"`
	}

	var structToEncode = SourceStruct{FName: "ben", LName: "pritchard", AgeInLife: 42}

	writerSchema := SchemaOf(&structToEncode)

	var encodedData bytes.Buffer

	err := writerSchema.Encode(&encodedData, structToEncode)
	if err != nil {
		t.Error(err)
		return
	}

	// write out our schema as JSON
	binarywriterSchema, err := writerSchema.MarshalJSON()
	if err != nil {
		t.Error("cannot create ", err)
	}

	// recreate the schmer from the JSON
	readerSchema, err := DecodeJSONSchema(binarywriterSchema)
	if err != nil {
		t.Error("cannot create writerSchema from raw JSON data", err)
		return
	}

	type DestinationStruct struct {
		FirstName string //`schemer:"FName"`
		LastName  string //`schemer:"LName"`
		Age       int    //`schemer:"AgeInLife"`
	}

	var structToDecode = DestinationStruct{}
	r := bytes.NewReader(encodedData.Bytes())

	err = readerSchema.Decode(r, &structToDecode)
	if err != nil {
		t.Error(err)
		return
	}

	// and now make sure that the structs match!
	decodeOK := true
	decodeOK = (structToDecode.FirstName == structToEncode.FName)
	decodeOK = decodeOK && (structToDecode.LastName == structToEncode.LName)
	decodeOK = decodeOK && (structToDecode.Age == structToEncode.AgeInLife)

	log.Print(structToDecode)

	if !decodeOK {
		t.Error("unexpected struct to struct decode")
	}

}

// TestDecodeFixedObject6 tests our ability to decode to an emptry interface
func TestDecodeFixedObject6(t *testing.T) {

	type SourceStruct struct {
		FName     string
		LName     string
		AgeInLife int
	}

	var structToEncode = SourceStruct{FName: "ben", LName: "pritchard", AgeInLife: 42}

	writerSchema := SchemaOf(&structToEncode)

	var encodedData bytes.Buffer

	err := writerSchema.Encode(&encodedData, structToEncode)
	if err != nil {
		t.Error(err)
	}

	var decodeDestination interface{}
	r := bytes.NewReader(encodedData.Bytes())

	err = writerSchema.Decode(r, &decodeDestination)
	if err != nil {
		t.Error(err)
	}

	v := reflect.ValueOf(decodeDestination).Elem()

	// and now make sure that the structs match!
	decodeOK := true
	decodeOK = (v.Field(0).String() == structToEncode.FName)
	decodeOK = decodeOK && (v.Field(1).String() == structToEncode.LName)
	decodeOK = decodeOK && (v.Field(2).Int() == int64(structToEncode.AgeInLife))

	if !decodeOK {
		t.Error("unexpected struct to struct decode")
	}

}
