package schemer

import (
	"bytes"
	"fmt"
	"testing"
)

func TestDecodeVarObject1(t *testing.T) {

	// setup an example schema
	varObjectSchema := VarObjectSchema{IsNullable: false}

	// encode it
	b := varObjectSchema.Bytes()

	// make sure we can successfully decode it
	var decodedIntSchema VarObjectSchema
	var err error

	tmp, err := NewSchema(b)
	if err != nil {
		t.Error("cannot encode binary encoded VarObjectSchema")
	}

	decodedIntSchema = tmp.(VarObjectSchema)

	// and then check the actual contents of the decoded schema
	// to make sure it contains the correct values
	if decodedIntSchema.IsNullable != varObjectSchema.IsNullable {
		t.Error("unexpected values when decoding binary EnumSchema")
	}

}

func TestDecodeVarObject2(t *testing.T) {

	strToIntMap := map[string]int{
		"rsc": 3711,
		"r":   2138,
		"gri": 1908,
		"adg": 912,
	}

	var buf bytes.Buffer
	var err error

	buf.Reset()

	// build up schema programatically...

	varObjectSchema := CreateVarObjectSchema(true)

	fixedIntSchema := CreateFixedIntegerSchema(true, 64, true)
	VarLenStringSchema := CreateVarLenStringSchema(true)

	of1 := VarObjectField{VarLenStringSchema, fixedIntSchema}

	varObjectSchema.Fields = append(varObjectSchema.Fields, of1)
	varObjectSchema.Fields = append(varObjectSchema.Fields, of1)
	varObjectSchema.Fields = append(varObjectSchema.Fields, of1)
	varObjectSchema.Fields = append(varObjectSchema.Fields, of1)

	//varObjectSchema := SchemaOf(&strToIntMap)

	err = varObjectSchema.Encode(&buf, strToIntMap)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("map to map")

	r := bytes.NewReader(buf.Bytes())

	mapToDecode := make(map[string]int, 4)

	err = varObjectSchema.Decode(r, &mapToDecode)
	if err != nil {
		t.Error(err)
	}

	// check each field...

}
