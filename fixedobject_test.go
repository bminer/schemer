package schemer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
)

func TestDecodeFixedObject1(t *testing.T) {

	type TestStruct struct {
		A string
		B string
		C [10]int
	}

	var testStruct TestStruct

	// setup an example schema
	fixedObjectSchema := SchemaOf(testStruct)

	// encode it
	b := fixedObjectSchema.Bytes()

	tmp, err := NewSchema(b)
	if err != nil {
		t.Error("cannot encode binary encoded FixedObjectSchema")
	}

	decodedIntSchema := tmp.(*FixedObjectSchema)

	// and then check the actual contents of the decoded schema
	// to make sure it contains the correct values
	if decodedIntSchema.IsNullable != fixedObjectSchema.Nullable() {
		t.Error("unexpected values when decoding binary FixedObjectSchema")
	}

}

func TestDecodeFixedObject2(t *testing.T) {

	type TestStructToNest2 struct {
		A [3]int8
		B []string
		C *string
		D **int
	}

	type TestStructToNest struct {
		Test123 float64
		M       map[string]int
		J       TestStructToNest2
	}

	type TestStruct struct {
		A string
		T TestStructToNest
	}

	i := 10
	var iptr *int = &i

	var structToEncode = TestStruct{"ben", TestStructToNest{99.9, map[string]int{
		"a": 1,
		"b": 2,
		"d": 3},
		TestStructToNest2{[3]int8{4, 5, 6}, []string{"go", "reflection", "rocks"}, nil, &iptr}}}

	var buf bytes.Buffer
	var err error

	buf.Reset()

	fixedObjectSchema := SchemaOf(&structToEncode)

	err = fixedObjectSchema.Encode(&buf, structToEncode)
	if err != nil {
		t.Error(err)
	}

	log.Print(buf.Bytes())

	fmt.Println("struct to struct")

	r := bytes.NewReader(buf.Bytes())

	var structToDecode = TestStruct{}

	log.Print(structToEncode)

	err = fixedObjectSchema.DecodeValue(r, reflect.ValueOf(&structToDecode))
	if err != nil {
		t.Error(err)
	}

	log.Print(structToDecode)

	/*
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(structToEncode, structToDecode) {
			t.Error("unexpected struct to struct decode")
		}

	*/

}

func TestDecodeFixedObject4(t *testing.T) {

	type TestStruct struct {
		A string
	}

	var structToEncode *TestStruct

	var buf bytes.Buffer
	var err error

	fixedObjectSchema := SchemaOf(&structToEncode)

	err = fixedObjectSchema.Encode(&buf, structToEncode)
	if err != nil {
		t.Error(err)
	}

	r := bytes.NewReader(buf.Bytes())

	var structToDecode = TestStruct{}

	err = fixedObjectSchema.DecodeValue(r, reflect.ValueOf(&structToDecode))

	if err != nil {
		t.Error(err)
	}

}

func TestDecodeFixedObject3(t *testing.T) {

	type TestStructToNest struct {
		Test123 float64
		M       map[string]int
	}

	type TestStruct struct {
		A string
		T *TestStructToNest
	}

	var structToEncode = TestStruct{}
	structToEncode.A = "ben"
	structToEncode.T = nil

	//var buf bytes.Buffer
	//var err error

	_ = SchemaOf(&structToEncode)

	//err = fixedObjectSchema.Encode(&buf, structToEncode)
	//if err != nil {
	//t.Error(err)
	//}

	//r := bytes.NewReader(buf.Bytes())

	//var structToDecode = TestStruct{}

	//err = fixedObjectSchema.DecodeValue(r, reflect.ValueOf(&structToDecode))

	//structToDecode.T.Test123 = 3.14
	//structToDecode.T.M["a"] = 1

	//log.Print(structToDecode.T.Test123)
	//log.Print(structToDecode.T.M)

	/*
		if err != nil {
			t.Error(err)
		}
	*/

}

// TestDecodeFixedObject5 tests our ability to decode objects to other objects, using struct tags....
func TestDecodeFixedObject5(t *testing.T) {

	type SourceStruct struct {
		FName     string //`schemer:"FirstName"`
		LName     string //`schemer:"LastName"`
		AgeInLife int    //`schemer:"Age"`
	}

	var structToEncode = SourceStruct{FName: "ben", LName: "pritchard"}

	writerSchema := SchemaOf(&structToEncode)

	var encodedData bytes.Buffer

	err := writerSchema.Encode(&encodedData, structToEncode)
	if err != nil {
		t.Error(err)
	}

	type DestinationStruct struct {
		FirstName string `schemer:"FName"`
		LastName  string `schemer:"LName"`
		Age       int    `schemer:"AgeInLife"`
	}

	var structToDecode = DestinationStruct{}
	r := bytes.NewReader(encodedData.Bytes())

	err = writerSchema.Decode(r, &structToDecode)
	if err != nil {
		t.Error(err)
	}

	// question: if certain fields cannot be decoded, we just throw an error...
	//				but actually, some of the fields could have been decoded OK
	//				is this OK???
	if err != nil {
		t.Error(err)
	}

	// and now make sure that the structs match!
	decodeOK := true
	decodeOK = (structToDecode.FirstName == structToEncode.FName)
	decodeOK = decodeOK && (structToDecode.LastName == structToEncode.LName)
	decodeOK = decodeOK && (structToDecode.Age == int(structToEncode.AgeInLife))

	log.Print(structToDecode)

	if !decodeOK {
		t.Error("unexpected struct to struct decode")
	}

	log.Println()

}

func SaveToDisk(fileName string, rawBytes []byte) {
	err := ioutil.WriteFile(fileName, rawBytes, 0644)
	if err != nil {
		panic(err)
	}
}

func ReadFromDisk(fileName string) []byte {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func TestWriter(t *testing.T) {

	type SourceStruct struct {
		XXX       string `schemer:"FName"`
		LName     string `schemer:"LastName"`
		AgeInLife int    `schemer:"Age"`
	}

	var structToEncode = SourceStruct{XXX: "ben", LName: "pritchard", AgeInLife: 42}

	writerSchema := SchemaOf(&structToEncode)
	binaryReaderSchema := writerSchema.Bytes()

	var encodedData bytes.Buffer

	err := writerSchema.Encode(&encodedData, structToEncode)
	if err != nil {
		t.Error(err)
	}

	SaveToDisk("/tmp/test.schema", binaryReaderSchema)
	SaveToDisk("/tmp/test.data", encodedData.Bytes())

}

func TestReader(t *testing.T) {

	// different order
	// different names
	type DestinationStruct struct {
		Age      int    `schemer:"AgeInLife"`
		LastName string `schemer:"LName"`
		YYY      string `schemer:"FName"`
	}

	var structToDecode = DestinationStruct{}

	binarywriterSchema := ReadFromDisk("/tmp/test.schema")
	writerSchema, err := NewSchema(binarywriterSchema)
	if err != nil {
		t.Error("cannot create writerSchema from raw binary data")
	}

	encodedData := ReadFromDisk("/tmp/test.data")
	r := bytes.NewReader(encodedData)

	err = writerSchema.Decode(r, &structToDecode)
	if err != nil {
		t.Error(err)
	}

	log.Println(structToDecode)

}
func TestBoth(t *testing.T) {
	TestReader(t)
	TestWriter(t)
}
