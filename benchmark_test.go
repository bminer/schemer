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

type benchmarkEmbeddedStruct struct {
	Int1 int64
	Int2 *int64
	Int3 *int64
}

// benchmarkStruct is what the routines in this test package test
// note that it only contains the types that are common to all encoding
// libraries
type benchmarkStruct struct {
	IntField1 int
	IntField2 *int
	IntField3 *int

	Bool1 bool
	Bool2 bool
	Bool3 *bool

	Object1 benchmarkEmbeddedStruct
	Object2 *benchmarkEmbeddedStruct
	Object3 *benchmarkEmbeddedStruct

	Float1 float64
	Float2 *float64
	Float3 *float64

	String1 string
	String2 *string
	String3 *string

	Slice1 []string
	Slice2 *[]string
	Slice3 *[]string
}

var writerSchema Schema
var benchmark benchmarkStruct

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

	benchmark.IntField1 = 101
	benchmark.IntField2 = nil
	intVal := 102
	benchmark.IntField3 = &intVal

	benchmark.Bool1 = false
	benchmark.Bool2 = false // nil
	tmpBool := false
	benchmark.Bool3 = &tmpBool

	benchmark.Object1 = benchmarkEmbeddedStruct{Int1: 3, Int2: nil}
	benchmark.Object2 = nil
	tmpStruct := benchmarkEmbeddedStruct{Int1: 4}
	benchmark.Object3 = &tmpStruct
	var tmpInt2 int64 = 5
	benchmark.Object3.Int3 = &tmpInt2

	benchmark.Float1 = 3.14
	benchmark.Float2 = nil
	tmpFloat := 99.55
	benchmark.Float3 = &tmpFloat

	benchmark.String1 = "hello, world"
	benchmark.String2 = nil
	tmpStr := "go reflection"
	benchmark.String3 = &tmpStr

	benchmark.Slice1 = []string{"a", "b", "c"}
	benchmark.Slice2 = nil
	tmpSlice := []string{"d", "e", "f"}
	benchmark.Slice3 = &tmpSlice

}

func schemerEncode() int {

	schemerData.Reset()

	err := writerSchema.Encode(&schemerData, benchmark)
	if err != nil {
		panic(err)
	}
	return len(schemerData.Bytes())
}

func schemerDecode() {
	var structToDecode = benchmarkStruct{}

	r := bytes.NewReader(schemerData.Bytes())

	err := writerSchema.Decode(r, &structToDecode)

	if err != nil {
		panic(err)
	}
}

func JSONEncode() int {
	var err error
	jsonData, err = json.Marshal(benchmark)
	if err != nil {
		panic(err)
	}
	return len(jsonData)
}

func JSONDecode() {
	var err error
	var structToDecode = benchmarkStruct{}
	err = json.Unmarshal(jsonData, &structToDecode)
	if err != nil {
		fmt.Printf("length is: %d", len(jsonData))
		panic(err)
	}

}

func GOBEncode() int {
	var err error
	enc := gob.NewEncoder(&gobData)

	err = enc.Encode(benchmark)
	if err != nil {
		panic(err)
	}

	return len(gobData.Bytes())
}

func GOBDecode() {
	var structToDecode = benchmarkStruct{}

	decoder := gob.NewDecoder(&gobData)

	err := decoder.Decode(&structToDecode)
	if err != nil {
		panic(err)
	}

	gobData.Reset()

}

func MessagePackEncode() int {
	var err error

	msgPackData, err = msgpack.Marshal(benchmark)

	if err != nil {
		panic(err)
	}

	return len(msgPackData)
}

func MessagePackDecode() {

	var structToDecode = benchmarkStruct{}

	err := msgpack.Unmarshal(msgPackData, &structToDecode)
	if err != nil {
		panic(err)
	}

}

func XMLEncode() int {

	encoder := xml.NewEncoder(&xmlData)
	encoder.Indent("", "\t")
	err := encoder.Encode(&benchmark)
	if err != nil {
		panic(err)
	}

	return len(xmlData.Bytes())
}

func XMLDecode() {
	var structToDecode = benchmarkStruct{}

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
	writerSchema = SchemaOf(&benchmark)
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

	writerSchema = SchemaOf(&benchmark)
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
