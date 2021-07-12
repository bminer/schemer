package schemer

import (
	"bytes"
	"io/ioutil"
	"log"
	"reflect"
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

// decode to empty interface
func FixedObjectReader1(t *testing.T, useJSON bool) {

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

	var destInterface interface{}
	var p *interface{} = &destInterface

	err = writerSchema.Decode(r, p)
	if err != nil {
		t.Error(err)
	}

	// when we decoded to an empty interface, null pointers we converted into zero values
	if reflect.ValueOf(destInterface).Elem().FieldByName("IntField2").Int() != 0 {
		t.Error("unexpected null int decode")
	}

	// when we decoded to an empty interface, null pointers we converted into zero values
	if reflect.ValueOf(destInterface).Elem().FieldByName("Complex1").Complex() != 3+2i {
		t.Error("unexpected complex decode")
	}

}

func FixedObjectReader2(t *testing.T, useJSON bool) {

	type destStruct struct {
		IntField2 int
	}

	var schemaFileName string
	var writerSchema Schema
	var err error
	var destStructToDecode destStruct

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

	err = writerSchema.Decode(r, &destStructToDecode)
	if err != nil {
		t.Error(err)
	}

	// when we decoded a nil value to a struct with out a pointer, we should have encoded
	// the zero value
	if destStructToDecode.IntField2 != 0 {
		t.Error("unexpected null int decode")
	}

}

func TestFixedObjectSerializeBinary(t *testing.T) {

	FixedObjectWriter(t, false)

	// test reading into different destinations
	FixedObjectReader1(t, false)
	FixedObjectReader2(t, false)

}

func TestFixedObjectSerializeJSON(t *testing.T) {

	FixedObjectWriter(t, true)
	FixedObjectReader1(t, true)

}
