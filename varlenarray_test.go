package schemer

import (
	"testing"
)

// make sure we can encode/decode binary schemas for EnumSchema
func TestDecodeVarLenArray1(t *testing.T) {

	// setup an example schema
	varLenArraySchema := VarLenArraySchema{IsNullable: false}

	// encode it
	b := varLenArraySchema.Bytes()

	// make sure we can successfully decode it
	var decodedIntSchema VarLenArraySchema
	var err error

	tmp, err := NewSchema(b)
	if err != nil {
		t.Error("cannot encode binary encoded VarLenArraySchema")
	}

	decodedIntSchema = tmp.(VarLenArraySchema)

	// and then check the actual contents of the decoded schema
	// to make sure it contains the correct values
	if decodedIntSchema.IsNullable != varLenArraySchema.IsNullable {
		t.Error("unexpected values when decoding binary EnumSchema")
	}

}
