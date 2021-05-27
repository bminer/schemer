package schemer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
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
	b := fixedObjectSchema.Bytes()

	tmp, err := DecodeSchema(b)
	if err != nil {
		t.Error("cannot encode binary encoded FixedObjectSchema")
	}

	decodedSchema := tmp.(*FixedObjectSchema)

	// and then check the actual contents of the decoded schema
	// to make sure it contains the correct values
	if decodedSchema.SchemaOptions.Nullable != fixedObjectSchema.Nullable() {
		t.Error("unexpected values when decoding binary FixedObjectSchema")
	}

}

// TestDecodeFixedObject2 tests encoding a nil string value
func TestDecodeFixedObject2(t *testing.T) {

	type StructToEncode struct {
		XX *string `schemer:""`
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
		YY *string `schemer:"[XX]"`
	}

	var structToDecode = StructToDecode{}

	fmt.Println(structToEncode)

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
		FName     string //`schemer:"FirstName"`
		LName     string //`schemer:"LastName"`
		AgeInLife int    //`schemer:"Age"`
	}

	var structToEncode = SourceStruct{FName: "ben", LName: "pritchard", AgeInLife: 42}

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

func saveToDisk(fileName string, rawBytes []byte) {
	err := ioutil.WriteFile(fileName, rawBytes, 0644)
	if err != nil {
		panic(err)
	}
}

func readFromDisk(fileName string) []byte {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func FixedObjectWriter(t *testing.T, useJSON bool) {

	type EmbeddedStruct struct {
		Int1 int64
		Int2 *int64
	}

	type SourceStruct struct {
		IntField1 int `schemer:"New_IntField1"`
		IntField2 *int

		Map1 map[string]bool `schemer:"New_Map1"`
		Map2 *map[string]bool

		Bool1 bool `schemer:"New_Bool1"`
		Bool2 *bool

		Complex1 complex64 `schemer:"New_Complex1"`
		Complex2 *complex64

		Array1 [5]string `schemer:"New_Array1"`
		Array2 *[5]string

		Object1 EmbeddedStruct `schemer:"New_Object1"`
		Object2 *EmbeddedStruct

		Float1 float64 `schemer:"New_Float1"`
		Float2 *float64

		String1 string `schemer:"New_String1"`
		String2 *string

		Slice1 []string `schemer:"New_Slice1"`
		Slice2 *[]string
	}

	// encode a nill value for field1
	var structToEncode SourceStruct

	structToEncode.IntField1 = 101
	structToEncode.IntField2 = nil

	structToEncode.Map1 = map[string]bool{"A": true, "B": false}
	structToEncode.Map2 = nil

	structToEncode.Bool1 = true
	structToEncode.Bool2 = nil

	structToEncode.Complex1 = 3 + 2i
	structToEncode.Complex2 = nil

	structToEncode.Array1 = [5]string{"1", "2", "3", "4", "5"}
	structToEncode.Array2 = nil

	structToEncode.Object1 = EmbeddedStruct{Int1: 3, Int2: nil}
	structToEncode.Object2 = nil

	structToEncode.Float1 = 3.14
	structToEncode.Float2 = nil

	structToEncode.String1 = "hello, world"
	structToEncode.String2 = nil

	structToEncode.Slice1 = []string{"a", "b", "c"}
	structToEncode.Slice2 = nil

	writerSchema := SchemaOf(&structToEncode)

	var binaryWriterSchema []byte
	var err error

	if useJSON {
		binaryWriterSchema, err = writerSchema.MarshalJSON()

		if err != nil {
			t.Error(err)
		}

	} else {
		binaryWriterSchema = writerSchema.Bytes()
	}

	var encodedData bytes.Buffer

	err = writerSchema.Encode(&encodedData, structToEncode)
	if err != nil {
		t.Error(err)
	}

	var schemaFileName string

	if useJSON {
		schemaFileName = "/tmp/test.schema"
	} else {
		schemaFileName = "/tmp/test.schema.json"
	}

	saveToDisk(schemaFileName, binaryWriterSchema)
	saveToDisk("/tmp/test.data", encodedData.Bytes())

}

func FixedObjectReader(t *testing.T, useJSON bool) {

	type EmbeddedStruct struct {
		Int1 int
		Int2 *int
	}

	type DestinationStruct struct {
		New_IntField1 int
		New_IntField2 *int `schemer:"IntField2"`

		New_Map1 map[string]bool
		New_Map2 *map[string]bool `schemer:"Map2"`

		New_Bool1 bool
		New_Bool2 *bool `schemer:"Bool2"`

		New_Complex1 complex64
		New_Complex2 *complex64 `schemer:"Complex2"`

		New_Array1 [5]string
		New_Array2 *[5]string `schemer:"Array2"`

		New_Object1 EmbeddedStruct
		New_Object2 *EmbeddedStruct `schemer:"Object2"`

		New_Float1 float64
		New_Float2 *float64 `schemer:"Float2"`

		New_String1 string
		New_String2 *string `schemer:"String2"`

		New_Slice1 []string
		New_Slice2 *[]string `schemer:"Slice2"`
	}

	var structToDecode = DestinationStruct{}
	var schemaFileName string
	var writerSchema Schema
	var err error

	if useJSON {
		schemaFileName = "/tmp/test.schema"
	} else {
		schemaFileName = "/tmp/test.schema.json"
	}

	binarywriterSchema := readFromDisk(schemaFileName)

	if useJSON {
		writerSchema, err = DecodeJSONSchema(binarywriterSchema)
		if err != nil {
			t.Error("cannot create writerSchema from raw JSON data")
		}
	} else {
		writerSchema, err = DecodeSchema(binarywriterSchema)
		if err != nil {
			t.Error("cannot create writerSchema from raw binary data")
		}
	}

	encodedData := readFromDisk("/tmp/test.data")
	r := bytes.NewReader(encodedData)

	err = writerSchema.Decode(r, &structToDecode)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(structToDecode)

}
func TestFixedObjectSerializeBinary(t *testing.T) {

	FixedObjectWriter(t, true)
	FixedObjectReader(t, true)

}

func TestFixedObjectSerializeJSON(t *testing.T) {

	FixedObjectWriter(t, false)
	FixedObjectReader(t, false)

}
