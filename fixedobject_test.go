package schemer

import (
	"bytes"
	"fmt"
	"log"
	"testing"
)

type TestStructToNest struct {
	Test123 int
}

type TestStruct struct {
	Name string
	Age  int
	I    [2]int
	T    TestStructToNest
}

func TestDecodeObject1(t *testing.T) {

	var structToEncode = TestStruct{"Ben", 13, [2]int{3, 5}, TestStructToNest{8}}

	var buf bytes.Buffer
	var err error

	buf.Reset()

	/*
		// build up schema programatically...

		stringSchema := CreateVarLenStringSchema(false)
		fixedIntSchema := CreateFixedIntegerSchema(true, 64, true)

		of1 := ObjectField{"Name", stringSchema}
		of2 := ObjectField{"Age", fixedIntSchema}

		fixedObjectSchema := FixedObjectSchema{}
		fixedObjectSchema.Fields = append(fixedObjectSchema.Fields, of1)
		fixedObjectSchema.Fields = append(fixedObjectSchema.Fields, of2)
	*/

	fixedObjectSchema := SchemaOf(&structToEncode)

	err = fixedObjectSchema.Encode(&buf, structToEncode)
	if err != nil {
		t.Error(err)
	}

	log.Print(buf.Bytes())

	fmt.Println("struct to struct")

	r := bytes.NewReader(buf.Bytes())

	structToDecode := TestStruct{"", 0, [2]int{0, 0}, TestStructToNest{0}}

	err = fixedObjectSchema.Decode(r, &structToDecode)
	if err != nil {
		t.Error(err)
	}

	// check each field...

}
