package schemer

import (
	"bytes"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
)

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

func populatestruct() {

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

	structToEncode.Complex1 = 3 + 2i
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

func fixedObjectWriter(t *testing.T, useJSON bool) {

	s, err := SchemaOf(&structToEncode)
	if err != nil {
		t.Error(err)
		return
	}

	writerSchema := s.(*FixedObjectSchema)

	var binaryWriterSchema []byte

	if useJSON {
		binaryWriterSchema, err = writerSchema.MarshalJSON()

		if err != nil {
			t.Error(err)
			return
		}

	} else {
		binaryWriterSchema, err = writerSchema.MarshalSchemer()
		if err != nil {
			t.Error(err)
			return
		}
	}

	var encodedData bytes.Buffer

	err = writerSchema.Encode(&encodedData, structToEncode)
	if err != nil {
		t.Error(err)
	}

	schemaFileName := "/tmp/test.schema"
	if useJSON {
		schemaFileName = schemaFileName + ".json"
	}

	saveToDisk(schemaFileName, binaryWriterSchema)
	saveToDisk("/tmp/test.data", encodedData.Bytes())

}

// decode to empty interface
func fixedObjectReader1(t *testing.T, useJSON bool) {

	var writerSchema Schema
	var err error

	schemaFileName := "/tmp/test.schema"
	if useJSON {
		schemaFileName = schemaFileName + ".json"
	}

	binarywriterSchema := readFromDisk(schemaFileName)

	if useJSON {
		writerSchema, err = DecodeSchemaJSON(bytes.NewReader(binarywriterSchema))
		if err != nil {
			t.Fatal("cannot create writerSchema from raw JSON data", err)
		}
	} else {
		writerSchema, err = DecodeSchema(bytes.NewReader(binarywriterSchema))
		if err != nil {
			t.Fatal("cannot create writerSchema from raw binary data", err)
		}
	}

	encodedData := readFromDisk("/tmp/test.data")
	r := bytes.NewReader(encodedData)

	var destInterface interface{}

	err = writerSchema.Decode(r, &destInterface)
	if err != nil {
		t.Error(err)
	}

	s := reflect.ValueOf(destInterface).Elem()

	// now check on some fields manually to make sure they decoded

	if s.FieldByName("IntField1").Int() != int64(structToEncode.IntField1) {
		t.Error("unexpected IntField1 decode")
	}

	// when we decoded to an empty interface, the encoded null int pointer should be decoded correctly to nil
	if !s.FieldByName("IntField2").IsNil() {
		t.Error("unexpected null int decode")
	}

	if s.FieldByName("IntField3").Elem().Int() != int64(*structToEncode.IntField3) {
		t.Error("unexpected null int decode")
	}

	// TODO check on other fields

	if s.FieldByName("Complex1").Complex() != 3+2i {
		t.Error("unexpected complex decode")
	}

}

// try to decode IntField2 into a struct that is not of pointer type...
// and then make sure Schemer throws the correct error
func fixedObjectReader2(t *testing.T, useJSON bool) {

	type destStruct struct {
		IntField2 int
	}

	var writerSchema Schema
	var err error
	var destStructToDecode destStruct

	schemaFileName := "/tmp/test.schema"
	if useJSON {
		schemaFileName = schemaFileName + ".json"
	}

	binarywriterSchema := readFromDisk(schemaFileName)

	if useJSON {
		writerSchema, err = DecodeSchemaJSON(bytes.NewReader(binarywriterSchema))
		if err != nil {
			t.Error("cannot create writerSchema from raw JSON data")
		}
	} else {
		writerSchema, err = DecodeSchema(bytes.NewReader(binarywriterSchema))
		if err != nil {
			t.Error("cannot create writerSchema from raw binary data")
		}
	}

	encodedData := readFromDisk("/tmp/test.data")
	r := bytes.NewReader(encodedData)

	log.Println("Checking (incorrect) decode of null pointer into non pointer type")
	err = writerSchema.Decode(r, &destStructToDecode)
	if err == nil {
		t.Error("unexpected successful decode")
	}
	log.Println("Schemer correctly refused to decode: ", err)
}

func TestFixedObjectSerializeBinary(t *testing.T) {

	populatestruct()
	fixedObjectWriter(t, false)

	// test reading into different destinations
	fixedObjectReader1(t, false)
	fixedObjectReader2(t, false)

}

func TestFixedObjectSerializeJSON(t *testing.T) {

	populatestruct()
	fixedObjectWriter(t, true)
	fixedObjectReader1(t, true)

}
