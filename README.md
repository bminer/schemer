# schemer

Lightweight and robust data encoding library for [Go](https://golang.org/)

Schemer provides an API to construct schemata that describe data structures; a schema is then used to encode and decode values into sequences of bytes to be sent over the network or written to a file.

Schemer seeks to be an alternative to [protobuf](https://github.com/protocolbuffers/protobuf) or [Avro](https://avro.apache.org/), but it can also be used as a substitute for [JSON](https://www.json.org/).

## Features

- Compact binary data format
- High-speed encoding and decoding
- Forward and backward compatibility
- No code generation and no [new language](https://en.wikipedia.org/wiki/Interface_description_language) to learn
- Simple and lightweight library with no external dependencies
- Supports custom encoding for user-defined data types
- JavaScript library for web browser interoperability (coming soon!)

## Why?

Schemer is an attempt to further simplify data encoding. Unlike other encoding libraries that use [interface description languages](https://en.wikipedia.org/wiki/Interface_description_language) (i.e. protobuf), schemer allows developers to construct schemata programmatically with an API. Rather than generating code from a schema, a schema can be constructed from code. In Go, schemata can be generated from Go types using the reflection library. This subtlety adds a surprising amount of flexibility and extensibility to the encoding library.

Here's how schemer stacks up against other encoding formats:

| Property                               | JSON               | XML                | MessagePack        | Protobuf           | Thrift             | Avro               | Gob                | Schemer            |
| -------------------------------------- | ------------------ | ------------------ | ------------------ | ------------------ | ------------------ | ------------------ | ------------------ | ------------------ |
| Human-Readable                         | :heavy_check_mark: | :neutral_face:     | :x:                | :x:                | :x:                | :x:                | :x:                | :x:                |
| Support for Many Programming Languages | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :x:                | :heavy_check_mark: |
| Widely Adopted                         | :heavy_check_mark: | :heavy_check_mark: | :x:                | :heavy_check_mark: | :x:                | :x:                | :x:                | :x:                |
| Precise Encoding of Numbers            | :neutral_face:     | :x:                | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: |
| Binary Strings                         | :x:                | :x:                | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: |
| Compact Encoded Payload                | :x::x:             | :x::x:             | :x:                | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: |
| Fast Encoding / Decoding               | :x:                | :x:                | :heavy_check_mark: | :heavy_check_mark: | :grey_question:    | :neutral_face:     | :neutral_face:     | :grey_question:    |
| Backward Compatibility                 | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :neutral_face:     | :neutral_face:     | :heavy_check_mark: | :neutral_face:     | :heavy_check_mark: |
| Forward Compatibility                  | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :neutral_face:     | :neutral_face:     | :heavy_check_mark: | :neutral_face:     | :heavy_check_mark: |
| No Language To Learn                   | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :x:                | :x:                | :neutral_face:     | :heavy_check_mark: | :heavy_check_mark: |
| Schema Support                         | :neutral_face:     | :neutral_face:     | :question:         | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :x:                | :heavy_check_mark: |
| Supports Fixed-field Objects           | :x:                | :x:                | :x:                | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: |
| Works on Web Browser                   | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :neutral_face:     | :heavy_check_mark: | :x:                | :calendar: soon…   |

## Types

schemer uses type information provided by the schema to encode values. The following are all of the types that are supported:

- Integer
  - Can be signed or unsigned
  - Fixed-size or variable-size [^1]
  	- Fixed-size integers can be 8, 16, 32, 64, or 128 bits
- Floating-point number (32 or 64-bit)
- Complex number (64 or 128-bit)
- Boolean
- Enumeration
- String
	- Can support any encoding, including UTF-8 and binary
	- Fixed-size or variable-size [^2]
- Array
	- Fixed-size or variable-size
- Object w/fixed fields (i.e. struct)
- Object w/variable fields (i.e. map)
- Schema (i.e. a schemer schema)
- Dynamically-typed value (i.e. variant)
- User-defined types
	- A few common types are provided for representing timestamps, time durations, IP addresses, UUIDs, regular expressions, etc.

[^1]: By default, integer types are encoded as variable integers, as this format will most likely generate the smallest encoded values.
[^2]: By default, string types are encoded as variable-size strings. Fixed-size strings are padded with trailing null bytes / zeros.

## Schema JSON Specification

### Types

| Type                     | JSON Type Name | Additional Options                                           |
| ------------------------ | -------------- | ------------------------------------------------------------ |
| Fixed-size Integer       | int            | * `signed` - boolean indicating if integer is signed or unsigned<br />* `bits` - one of the following numbers indicating the size of the integer: 8, 16, 32, 64, 128, 256, 512, 1024 |
| Variable-size Integer    | int            | * `signed` - boolean indicating if integer is signed or unsigned<br />* `bits` - must be `null` or omitted |
| Floating-point Number    | float          | * `bits` - one of the following numbers indicating the size of the floating-point: 32, 64 |
| Complex Number           | complex        | * `bits` - one of the following numbers indicating the size of the complex number: 64, 128 |
| Boolean                  | bool           |                                                              |
| Enum                     | enum           | * `values` - an object mapping strings to integer values     |
| Fixed-Length String      | string         | * `length` - the length of the string in bytes               |
| Variable-Length String   | string         | * `length` - must be `null` or omitted                       |
| Fixed-Length Array       | array          | * `length` - the length of the string in bytes               |
| Variable-Length Array    | array          | * `length` - must be `null` or omitted                       |
| Object w/fixed fields    | object         | * `fields` - an array of fields. Each field is an type object with keys:<br />`name`[^3], `type`, and any additional options for the `type` |
| Object w/variable fields | object         | * `fields` - must be `null` or omitted                       |
| Variant                  | variant        |                                                              |

[^3]: It is strongly encouraged to use [camelCase](https://en.wikipedia.org/wiki/Camel_case) for object field names.

### Example

Here's a struct with three fields:

- firstName (string)
- lastName (string)
- age (uint8 - unsigned integer requiring a single byte)

```json
{
  "type": "object",
  "fields": [
    {
      "name": "firstName",
      "type": "string"
    }, {
      "name": "lastName",
      "type": "string"
    }, {
      "name": "age",
      "type": "int",
      "signed": false,
      "size": 1
    }
  ]
}
```

## Type Compatibility

When decoding values from one type to another, schemer employs the following compatibility rules. These rules, while rather opinionated, provide safe defaults when decoding values. Users who want to carefully craft how values are decoded from one type to another can simply create a custom type.

As a general rule, types are only compatible with themselves (i.e. boolean values can only be decoded to boolean values).  The table below outlines a few notable exceptions and describes how using "weak" decoding mode can increase type compatibility by sacrificing type safety and by making a few assumptions.

|                 | Destination           |                       |                        |                        |                       |                        |                        |                       |
| --------------- | --------------------- | --------------------- | ---------------------- | ---------------------- | --------------------- | ---------------------- | ---------------------- | --------------------- |
| **Source**      | int                   | float                 | complex                | bool                   | enum                  | string                 | array (see #12)        | object                |
| int             | :heavy_check_mark: #1 | :heavy_check_mark: #1 | :heavy_check_mark: #1  | :grey_exclamation: #6  | :grey_exclamation: #7 | :grey_exclamation: #9  | :x:                    | :x:                   |
| float           | :heavy_check_mark: #1 | :heavy_check_mark: #1 | :heavy_check_mark: #1  | :x:                    | :x:                   | :grey_exclamation: #9  | :x:                    | :x:                   |
| complex         | :heavy_check_mark: #1 | :heavy_check_mark: #1 | :heavy_check_mark: #1  | :x:                    | :x:                   | :grey_exclamation: #9  | :grey_exclamation: #11 | :x:                   |
| bool            | :grey_exclamation: #6 | :x:                   | :x:                    | :heavy_check_mark:     | :x:                   | :grey_exclamation: #10 | :x:                    | :x:                   |
| enum            | :grey_exclamation: #7 | :x:                   | :x:                    | :x:                    | :heavy_check_mark: #2 | :heavy_check_mark: #2  | :x:                    | :x:                   |
| string          | :grey_exclamation: #8 | :grey_exclamation: #8 | :grey_exclamation: #8  | :grey_exclamation: #10 | :heavy_check_mark: #2 | :heavy_check_mark:     | :x:                    | :x:                   |
| array (see #12) | :x:                   | :x:                   | :grey_exclamation: #11 | :x:                    | :x:                   | :x:                    | :heavy_check_mark: #3  | :x:                   |
| object          | :x:                   | :x:                   | :x:                    | :x:                    | :x:                   | :x:                    | :x:                    | :heavy_check_mark: #4 |

**Legend**:<br/>:heavy_check_mark: - indicates compatibility according to the specified rule<br/>:grey_exclamation:- indicates compatibility according to the specified rule only if weak decoding is used<br/>:x: - indicates that the source type cannot be decoded to the destination

#### Compatibility Rules:

1. Any number can be decoded to any other number, provided the decoded value can be stored into the destination without losing any precision. If weak decoding is specified, we loosen this restriction slightly by allowing floating-point and complex number conversions to lose precision.

	For example, if the number `3.14` is decoded, it can be stored as a float or complex number, but it cannot be stored as an integer. Similarly, the number `500` can be stored into a `uint16` but not a `uint8`, since `uint8` can only store values between 0 and 255.

1. Enumerations are decoded to other enumerations by performing a case-insensitive match on the named value, not a match on the numeric value. If multiple matches occur, a case-sensitive match is then performed. Decoding fails if the decoded named value does not match a named value in the destination enumeration. Enumerations can also be converted to strings and vice-versa by matching on the enumeration's named value.

1. Arrays can be decoded to arrays if the element type and array length is compatible. Specifically, when the destination array is of fixed-size and does not support null values, the decoded array must match exactly in length.

1. Objects are decoded to other objects by performing a case-insensitive match on the key or field name.  If multiple matches occur, a case-sensitive match is then performed. When the destination is an object with fixed fields and the decoded value does not have a matching key or field name, the key / field is simply skipped and will remain unchanged.

1. Null values can only be decoded to destinations that support null values (i.e. pointers), but a non-null value can be decoded even if the destination does not support null values.

The following compatibility rules apply for weak decoding only:

6. The boolean value `true` can be converted to the integer value `1`, and the boolean value `false` can be converted to the integer value `0`. Similarly, the integer `0` will be decoded as `false`, and all other integers are decoded as `true`.
6. Enumerations can be converted to integer values and vice-versa, and they are matched on the enumeration's numeric value.
6. Strings can be decoded to numeric values by considering the string format according to the table below. The resulting numeric value is compatible with the destination according to the relevant compatibility rules.
6. Numbers are always encoded to strings in base 10.
6. Boolean values `true` and `false` are converted to string values `"true"` and `"false"` respectively. Strings `"1"`, `"t"`, `"T"`, `"TRUE"`, `"true"`, and `"True"` can be converted to the boolean value `true`. Strings `"0"`, `"f"`, `"F"`, `"FALSE"`, `"false"`, and `"False"` can be converted to boolean value `false`.
6. Complex numbers may be converted into 2-element arrays of floating-point numbers and vice-versa. The real part of the complex number will be matched with array element 0, and the complex part will be matched with array element 1.
6. Single-element arrays can be decoded to a destination that is compatible with the array element and vice-versa.

#### String to number decoding:

| String Example | Regular Expression | Decoded As              |
| -------------- | ------------------ | ----------------------- |
| `"-3.14"`      |                    | Number, base 10         |
| `"0b1101"`     |                    | Number, base 2          |
| `"0775"`       |                    | Number, base 8          |
| `"0x2020"`     |                    | Number, base 16         |
| `"2.34 + 2i"`  |                    | Complex number, base 10 |

## API

Schemata are responsible for encoding and decoding data. A schema can be defined programatically, or it can be generated using reflection.

```go
type Schema interface {
  // Encode uses the schema to write the encoded value of v to the output stream
  Encode(w io.Writer, v interface{}) error
  // Decode uses the schema to read the next encoded value from the input stream and store it in v
  Decode(r io.Reader, v interface{}) error
  // Bytes encodes the schema in a portable binary format
  Bytes() []byte
  // String returns the schema in a human-readable format
  // String() string
  // MarshalJSON returns the JSON encoding of the schema
  MarshalJSON() ([]byte, error)
  // UnmarshalJSON updates the schema by decoding the JSON-encoded schema in b
  UnmarshalJSON(b []byte) error
  // Nullable returns true if and only if the type is nullable
  Nullable() bool
  // SetNullable sets the nullable flag for the schema
  SetNullable(n bool)
}
type basicSchema struct {
  header byte
}
type enumSchema struct {
  Values map[int]string
}
type fixedStringSchema struct {
  Length int64
}
type fixedArraySchema struct {
  Length  int64
  Element Schema
}
type varArraySchema struct {
  Element Schema
}
type fixedObjectSchema struct {
  Fields []ObjectField
}
type objectField struct {
  Name   string
  Schema Schema
}
type varObjectSchema struct {
  Key   Schema
  Value Schema
}
// Functions to create a Schema
func FixedIntegerSchema(signed bool, bits int) Schema
func VarIntegerSchema(signed bool) Schema
func FloatSchema(bits int) Schema
func ComplexSchema(bits int) Schema
func BooleanSchema() Schema
// SchemaOf generates a Schema from the concrete value stored in the interface i.
// SchemaOf(nil) returns a Schema for an empty struct.
func SchemaOf(i interface{}) Schema
// SchemaOfType generates a Schema from t.
// SchemaOfType(nil) returns a Schema for an empty struct.
func SchemaOfType(t reflect.Type) Schema
// DecodeSchema decodes a schema stored in buf and returns an error if the schema is invalid
func DecodeSchema(buf []byte) (Schema, error)
```

The Encoder and Decoder use buffers under the hood to avoid numerous I/O calls to the underlying streams.

```go
type Decoder struct {
  // unexported fields
}
// NewDecoder returns a new decoder that reads encoded values from r
// using the given schema.
func NewDecoder(r io.Reader, s Schema) *Decoder
// Decode uses the schema to read the next value from the input stream
// and store it in the data represented by the empty interface
func (dec *Decoder) Decode(e interface{}) error
```

```go
type Encoder struct {
  // unexported fields
}
// NewEncoder returns a new encoder that writes encoded values to w using the
// given schema.
func NewEncoder(w io.Writer, s Schema) *Encoder
// Encode uses the schema to write the encoded value of v to the stream.
// Returns an error if v cannot be encoded using the schema or if an I/O error occurs
func (enc *Encoder) Encode(v interface{}) error
```

With Go Generics, typed encoders and decoders are possible.  One advantage here is the ability to validate the compatibility between the type parameter and the schema in advance, although I'm not sure that provides any practical benefit.

```go
type TypedSchema[T] interface {
    // Validate returns true if and only if T can be encoded and decoded by this Schema
    Validate() bool
    // Encode uses the schema to write the encoded value of v to the output stream
    Encode(w io.Writer, v T)
    // Decode uses the schema to read the next encoded value from the input stream and store it in v
    Decode(r io.Reader, v T)
    // Bytes encodes the schema in a portable binary format
    Bytes() []byte
    // String returns the schema in a human-readable format
    String() string
    // MarshalJSON returns the JSON encoding of the schema
    MarshalJSON() ([]byte, error)
    // UnmarshalJSON updates the schema decodes the JSON encoded schema
    UnmarshalJSON(b []byte) error
}
type TypedDecoder[T] struct {}
func NewTypedDecoder[T](r io.Reader, s TypedSchema[T]) *TypedDecoder[T]
func (dec TypedDecoder[T]) Decode(v T) error
type TypedEncoder[T] struct {}
func NewTypedEncoder[T](w io.Writer, s TypedSchema[T]) *TypedEncoder[T]
func (enc TypedEncoder[T]) Encode(v T) error
```

This library was created on April 14, 2021, the day of Bernie Madoff's death. May he rest in peace.