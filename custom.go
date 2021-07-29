package schemer

import "reflect"

type CustomSchema interface {

	// returns the name of the schema
	Name() string

	// returns the UUID of the custom schema
	UUID() byte

	// returns a schema unmarshalled from the passed in JSON data
	Unmarshaljson(buf []byte) (Schema, error)

	// returns a schema unmarshalled from the passed in binary data
	UnMarshalSchemer(buf []byte, byteIndex *int) (Schema, error)

	// returns a schema if the passed in reflect.type is handled by this custom schema
	ForType(t reflect.Type) Schema
}

// global variable representing all registered schemas
var registeredSchemas []CustomSchema

func RegisteredSchemas() []CustomSchema {
	return registeredSchemas
}

// registers a custom schema
func RegisterCustomSchema(s CustomSchema) {
	registeredSchemas = append(registeredSchemas, s)
}
