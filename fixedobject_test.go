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
		B string
		C [10]int
	}

	var testStruct TestStruct

	// setup an example schema
	fixedObjectSchema := SchemaOf(testStruct).(FixedObjectSchema)

	// encode it
	b := fixedObjectSchema.Bytes()

	// make sure we can successfully decode it
	var decodedIntSchema FixedObjectSchema
	var err error

	tmp, err := NewSchema(b)
	if err != nil {
		t.Error("cannot encode binary encoded FixedObjectSchema")
	}

	decodedIntSchema = tmp.(FixedObjectSchema)

	// and then check the actual contents of the decoded schema
	// to make sure it contains the correct values
	if decodedIntSchema.IsNullable != fixedObjectSchema.IsNullable {
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

	log.Print(**structToDecode.T.J.D)

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

	type WriterSchema struct {
		FName     string `schemer:"FirstName"`
		LName     string `schemer:"LastName"`
		AgeInLife int    `schemer:"Age"`
	}

	type DestinationSchema struct {
		FirstName string
		LastName  string
		Age       uint8
	}

	var structToEncode = WriterSchema{FName: "ben", LName: "pritchard", AgeInLife: 42}

	fixedObjectSchema := SchemaOf(&structToEncode).(FixedObjectSchema)

	var err error
	var buf bytes.Buffer

	err = fixedObjectSchema.Encode(&buf, structToEncode)
	if err != nil {
		t.Error(err)
	}

	r := bytes.NewReader(buf.Bytes())

	var structToDecode = DestinationSchema{}

	err = fixedObjectSchema.DecodeValue(r, reflect.ValueOf(&structToDecode))
	if err != nil {
		t.Error(err)
	}

	// and now make sure that the structs match!
	decodeOK := true
	decodeOK = (structToDecode.FirstName == structToEncode.FName)
	decodeOK = decodeOK && (structToDecode.LastName == structToEncode.LName)
	//decodeOK = decodeOK && (structToDecode.Age == int(structToEncode.AgeInLife))

	if !decodeOK {
		t.Error("unexpected struct to struct decode")
	}

	log.Println()

}
