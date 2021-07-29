package schemer

import (
	"fmt"
)

type CustomSchema interface {

	// returns the name of the
	Name() string

	// returns the UUID of the custom schema
	UUID() byte

	// returns a schema unmarshalled from the passedin JSON data
	UnMarshalJSON(buf []byte) (Schema, error)

	// returns a schema unmarshalled from the passed in binary data
	UnMarshalSchemer(buf []byte, byteIndex *int) (Schema, error)
}

var RegisteredSchemas []CustomSchema

// must have some sort of global
func RegisterCustomSchema(s CustomSchema) {
	fmt.Printf("%s schema registered!", s.Name())
	RegisteredSchemas = append(RegisteredSchemas, s)

}
