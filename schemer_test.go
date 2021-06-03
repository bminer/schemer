package schemer

// schmer_test is the main testing package for this library that is designed to test the overall functionality of
// schemer. The tests are designed to include encoding/decoding all of the types that we support in schemer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"testing"
)

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

	type embeddedStruct struct {
		Int1 int64
		Int2 *int64
		Int3 *int64
	}

	// sourceStruct is what the routines in this test package
	type sourceStruct struct {
		IntField1 int
		IntField2 *int
		IntField3 *int

		Map1 map[string]bool
		Map2 *map[string]bool

		Map3 map[string]bool

		Bool1 bool
		Bool2 *bool
		Bool3 *bool

		Complex2 *complex64
		Complex3 *complex64

		Complex4 complex128
		Complex5 *complex128
		Complex6 *complex128

		Array1 [5]string
		Array2 *[5]string
		Array3 *[5]string

		Object1 embeddedStruct
		Object2 *embeddedStruct
		Object3 *embeddedStruct

		Float1 float64
		Float2 *float64
		Float3 *float64

		String1 string
		String2 *string
		String3 *string

		Slice1 []string
		Slice2 *[]string
		Slice3 *[]string

		Complex1 complex64
	}

	// encode a nill value for field1
	var structToEncode sourceStruct

	structToEncode.IntField1 = 101
	structToEncode.IntField2 = nil
	intVal := 102
	structToEncode.IntField3 = &intVal

	structToEncode.Map1 = map[string]bool{"A": true, "B": false}
	structToEncode.Map2 = nil
	tmpMap := map[string]bool{"C": false, "D": true}
	structToEncode.Map3 = tmpMap

	structToEncode.Bool1 = false
	structToEncode.Bool2 = nil
	tmpBool := false
	structToEncode.Bool3 = &tmpBool

	structToEncode.Complex1 = 3 + 2i
	structToEncode.Complex2 = nil
	var tmpComplex1 complex64 = 4 + 5i
	structToEncode.Complex3 = &tmpComplex1

	structToEncode.Complex4 = 5 + 6i
	structToEncode.Complex5 = nil
	var tmpComplex2 complex128 = 7 + 8i
	structToEncode.Complex6 = &tmpComplex2

	structToEncode.Array1 = [5]string{"1", "2", "3", "4", "5"}
	structToEncode.Array2 = nil
	tmpArray := [5]string{"6", "7", "8", "9", "10"}
	structToEncode.Array3 = &tmpArray

	structToEncode.Object1 = embeddedStruct{Int1: 3, Int2: nil}
	structToEncode.Object2 = nil
	tmpStruct := embeddedStruct{Int1: 4}
	structToEncode.Object3 = &tmpStruct
	var tmpInt2 int64 = 5
	structToEncode.Object3.Int3 = &tmpInt2

	structToEncode.Float1 = 3.14
	structToEncode.Float2 = nil
	tmpFloat := 99.55
	structToEncode.Float3 = &tmpFloat

	structToEncode.String1 = "hello, world"
	structToEncode.String2 = nil
	tmpStr := "go reflection"
	structToEncode.String3 = &tmpStr

	structToEncode.Slice1 = []string{"a", "b", "c"}
	structToEncode.Slice2 = nil
	tmpSlice := []string{"d", "e", "f"}
	structToEncode.Slice3 = &tmpSlice

	writerSchema := SchemaOf(&structToEncode)

	var binaryWriterSchema []byte
	var err error

	if useJSON {
		binaryWriterSchema, err = writerSchema.MarshalJSON()

		if err != nil {
			t.Error(err)
		}

	} else {
		binaryWriterSchema = writerSchema.MarshalSchemer()
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

	type embeddedStruct struct {
		Int1 int
		Int2 *int
	}

	type DestinationStruct struct {

		/*
			New_IntField1 int
			New_IntField2 *int

			New_Map1 map[string]bool
			New_Map2 *map[string]bool

			New_Bool1 bool
			New_Bool2 *bool

			New_Complex1 complex64
			New_Complex2 *complex64

			New_Array1 [5]string
			New_Array2 *[5]string

			New_Object1 embeddedStruct
			New_Object2 *embeddedStruct

			New_Float1 float64
			New_Float2 *float64

			New_String1 string
			New_String2 *string

			New_Slice1 []string
			New_Slice2 *[]string
		*/

		Complex1 complex64
	}

	var structToDecode DestinationStruct
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

	if structToDecode.Complex1 != 3+2i {
		t.Error("unexpected complex decode")
	}

	fmt.Println(structToDecode)

}
func TestFixedObjectSerializeBinary(t *testing.T) {

	FixedObjectWriter(t, false)
	FixedObjectReader(t, false)

}

func TestFixedObjectSerializeJSON(t *testing.T) {

	FixedObjectWriter(t, true)
	FixedObjectReader(t, true)

}
