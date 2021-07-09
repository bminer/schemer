package schemer

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
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

	/*
		Map1 map[string]bool
		Map2 *map[string]bool

		Map3 map[string]bool
	*/

	Bool1 bool
	Bool2 bool
	Bool3 *bool

	/*
		Complex2 *complex64
		Complex3 *complex64

		Complex4 complex128
		Complex5 *complex128
		Complex6 *complex128
	*/

	/*
		Array1 [5]string
		Array2 *[5]string
		Array3 *[5]string
	*/

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

	//Complex1 complex64
}

var writerSchema Schema
var structToEncode sourceStruct

var schemerData bytes.Buffer
var jsonData []byte
var gobData bytes.Buffer
var msgPackData []byte

var xmlData bytes.Buffer

//var xmlData *bytes.Buffer = &bytes.Buffer{}

var printedAlready bool

// to run the benchmarks only:
// go test -run=XXX -bench.

func PopulateStructToEncode() {

	structToEncode.IntField1 = 101
	structToEncode.IntField2 = nil
	intVal := 102
	structToEncode.IntField3 = &intVal

	/*

		structToEncode.Map1 = map[string]bool{"A": true, "B": false}
		structToEncode.Map2 = nil
		tmpMap := map[string]bool{"C": false, "D": true}
		structToEncode.Map3 = tmpMap

	*/

	structToEncode.Bool1 = false
	structToEncode.Bool2 = false // nil
	tmpBool := false
	structToEncode.Bool3 = &tmpBool

	/*
		structToEncode.Complex1 = 3 + 2i
		structToEncode.Complex2 = nil
		var tmpComplex1 complex64 = 4 + 5i
		structToEncode.Complex3 = &tmpComplex1

		structToEncode.Complex4 = 5 + 6i
		structToEncode.Complex5 = nil
		var tmpComplex2 complex128 = 7 + 8i
		structToEncode.Complex6 = &tmpComplex2
	*/

	/*

		structToEncode.Array1 = [5]string{"1", "2", "3", "4", "5"}
		structToEncode.Array2 = nil
		tmpArray := [5]string{"6", "7", "8", "9", "10"}
		structToEncode.Array3 = &tmpArray

	*/

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

}

func schemerEncode() int {

	schemerData.Reset()

	err := writerSchema.Encode(&schemerData, structToEncode)
	if err != nil {
		panic(err)
	}
	return len(schemerData.Bytes())
}

func schemerDecode() {
	var structToDecode = sourceStruct{}

	r := bytes.NewReader(schemerData.Bytes())

	err := writerSchema.Decode(r, &structToDecode)

	if err != nil {
		panic(err)
	}
}

func JSONEncode() int {
	var err error
	jsonData, err = json.Marshal(structToEncode)
	if err != nil {
		panic(err)
	}
	return len(jsonData)
}

func JSONDecode() {
	var err error
	var structToDecode = sourceStruct{}
	err = json.Unmarshal(jsonData, &structToDecode)
	if err != nil {
		fmt.Printf("length is: %d", len(jsonData))
		panic(err)
	}

}

func GOBEncode() int {
	var err error
	enc := gob.NewEncoder(&gobData)

	err = enc.Encode(structToEncode)
	if err != nil {
		panic(err)
	}

	return len(gobData.Bytes())
}

func GOBDecode() {
	var structToDecode = sourceStruct{}

	decoder := gob.NewDecoder(&gobData)

	err := decoder.Decode(&structToDecode)
	if err != nil {
		panic(err)
	}

	gobData.Reset()

}

func MessagePackEncode() int {
	var err error

	msgPackData, err = msgpack.Marshal(structToEncode)

	if err != nil {
		panic(err)
	}

	return len(msgPackData)
}

func MessagePackDecode() {

	var structToDecode = sourceStruct{}

	err := msgpack.Unmarshal(msgPackData, &structToDecode)
	if err != nil {
		panic(err)
	}

}

func XMLEncode() int {

	encoder := xml.NewEncoder(&xmlData)
	encoder.Indent("", "\t")
	err := encoder.Encode(&structToEncode)
	if err != nil {
		panic(err)
	}

	return len(xmlData.Bytes())
}

func XMLDecode() {
	var structToDecode = sourceStruct{}

	decoder := xml.NewDecoder(&xmlData)
	_ = decoder.Decode(&structToDecode)

	/*
		if err != nil {
			panic(err)
		}
	*/

}

func BenchmarkSchemerEncode(b *testing.B) {
	PopulateStructToEncode()
	writerSchema = SchemaOf(&structToEncode)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		schemerEncode()
	}

}

func BenchmarkSchemerDecode(b *testing.B) {

	for n := 0; n < b.N; n++ {
		schemerDecode()
	}

}

func BenchmarkJSONEncode(b *testing.B) {

	PopulateStructToEncode()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		JSONEncode()
	}

}

func BenchmarkJSONDecode(b *testing.B) {

	for n := 0; n < b.N; n++ {
		JSONDecode()
	}

}

func BenchmarkGOBEncode(b *testing.B) {

	PopulateStructToEncode()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		GOBEncode()
	}
}

func _BenchmarkGOBDecode(b *testing.B) {

	for n := 0; n < b.N; n++ {
		GOBDecode()
	}

}

func BenchmarkMessagePackEncode(b *testing.B) {

	PopulateStructToEncode()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		MessagePackEncode()
	}

}

func BenchmarkMessagePackDecode(b *testing.B) {

	for n := 0; n < b.N; n++ {
		MessagePackDecode()
	}

}

func BenchmarkXMLEncodeb(b *testing.B) {

	PopulateStructToEncode()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		XMLEncode()
		XMLDecode()
	}

	//XMLDecode()

}

func BenchmarkXMLDecode(b *testing.B) {

	for n := 0; n < b.N; n++ {
		XMLDecode()
	}

}

func BenchmarkResults(b *testing.B) {

	writerSchema = SchemaOf(&structToEncode)
	b.ResetTimer()

	size1 := schemerEncode()
	size2 := JSONEncode()
	size3 := GOBEncode()
	size4 := MessagePackEncode()
	size5 := XMLEncode()

	if !printedAlready {

		fmt.Printf("Schemer data size......... %d\n", size1)
		fmt.Printf("JSON data size............ %d\n", size2)
		fmt.Printf("GOB data size............. %d\n", size3)
		fmt.Printf("MessagePack data size..... %d\n", size4)
		fmt.Printf("XML data size............. %d\n", size5)

	}

	printedAlready = true

}