package schemer

import (
	"testing"
)

// TestDecodeString1 checks that we can encode / decode the binary schema for a string
func TestDecodeString1(t *testing.T) {

	// setup an example schema
	schema := StringSchema{IsNullable: true, IsFixedLength: false, FixedLength: 80}

	// encode it
	b := schema.Bytes()

	// make sure we can successfully decode it
	var decodedStringSchema StringSchema
	var err error

	tmp, err := NewSchema(b)
	if err != nil {
		t.Error("cannot decode binary encoded string schema")
	}

	decodedStringSchema = tmp.(StringSchema)
	if schema.IsNullable != decodedStringSchema.IsNullable ||
		schema.IsFixedLength != decodedStringSchema.IsFixedLength {

		// nothing else to test here...
		// really this shouldn't ever happen

		t.Error("unexpected value for BoolSchema")
	}

}
