# schemer

Lightweight and robust data encoding library for [Go](https://golang.org/)

Schemer provides an API to construct schemata that describe data structures; the schema is then used to encode and decode values.

Schemer seeks to be an alternative to [protobuf](https://github.com/protocolbuffers/protobuf), but it can also be used as a substitute for [JSON](https://www.json.org/) or [XML](https://en.wikipedia.org/wiki/XML).

## Features

- Compact binary data format
- High-speed encoding and decoding
- Forward and backward compatibility
- No code generation and no new language to learn
- Simple and lightweight library with no external dependencies
- JavaScript library for web browser interoperability (coming soon!)

## Why?

[protobuf](https://github.com/protocolbuffers/protobuf) has several drawbacks that schemer addresses:

* protobuf uses [manually-assigned field identifiers](https://developers.google.com/protocol-buffers/docs/proto3#assigning_field_numbers) to ensure backward compatibility, but some may find this approach to be cumbersome and verbose. Much like [Avro](https://avro.apache.org/docs/current/), schemer ensures forward and backward compatibility by encoding values along with the writer's schema. During decoding, discrepancies between the reader schema and writer schema can be easily resolved.
* Schemer allows the schema and encoded values to be written separately; thus, each datum is written with no per-value overheads. This reduces the size of the encoded data and improves performance.
* Schemer relies on [Go's reflect package](https://golang.org/pkg/reflect/), so no code generation is needed. Schemata can be generated from vanilla Go data structures without writing a separate schema document. There is no new language to learn.

## Types

schemer uses type information provided by the schema to encode values. The following are all of the types that are supported:

- Number
  - Fixed-size Integer
  	- Can be signed or unsigned
  	- Integer sizes of 8, 16, 32, 64, 128 bits, although integers larger than 64 bits are rare.
  - Variable-size Integer [^1] (signed or unsigned)
  - Floating-point number (32 or 64-bit)
  - Complex number (64 or 128-bit)
- Enumeration
- Boolean
- Fixed-size String (UTF-8)
- Variable-size String (UTF-8)
- Fixed-size Array
- Variable-size Array
- Object w/fixed fields (i.e. struct)
- Object w/variable fields (i.e. map)
- Dynamically-typed value (i.e. variant)
- User-defined types using hooks
	- Common hooks are provided (i.e. `time.Time`, `net.IPAddr`, etc.)

## Schema Format

Here's how schemer encodes the type information into a single byte in the schema:

| Type                   | Encoded Type Byte | Notes                                                        |
| ---------------------- | ----------------- | ------------------------------------------------------------ |
| Fixed-size Integer | 0b0000 nnns      | where s is the signed/unsigned bit and n represents the encoded integer size in (8 << n) bits. Example: 0x07 is a signed 64-bit integer. |
| Variable-size Integer[^1] | 0b0001 000s       | where s is the signed/unsigned bit                         |
| Floating-point Number  | 0b0001 01*n       | where n is the floating-point size in (32 << n) bits and * is reserved for future use |
| Complex Number        | 0b0001 10*n       | where n is the complex number size in (64 << n) bits and * is reserved for future use |
| Enum | 0b0001 1100 |  |
| Boolean            | 0b0001 1110 |                                                              |
| String                 | 0b0010 000f       | where f indicates that the string is of fixed byte length    |
| Array                  | 0b0010 010f     | where f indicates that the array is of fixed length          |
| Object                 | 0b0010 100f   | where f indicates that the object has fixed number of fields |
| Variant | 0b0010 1100 |  |
| Nullable [^2]          | 0b1xxx xxxx       | where x is the type information for the nullable type |

[^1]: By default, integer types are encoded as variable integers, as this format will most likely generate the smallest encoded size
[^2]: This highest bit of the encoded type byte is the nullable bit, indicating whether or not the value can be null.  If set, the encoded value can also be null.

## Values

The following describes how schemer encodes different values.

| Type                     | Encoding Format                                              | Nullable Format [^3]                                         |
| ------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
| Fixed-size Integer       | Exactly `1 >> n` bytes of two's complement integer representation in little-endian byte order | Preceded by 1 byte. 0 indicates not null.                    |
| Variable-size Integer    | Each byte's most significant bit indicates more bytes follow. Lower 7 bits are concatenated (in big-endian order) to form [ZigZag-encoded integer](https://developers.google.com/protocol-buffers/docs/encoding?csw=1#types). | Initial byte can only store 6-bit value, and most significant bit indicates null if set. |
| Floating-point Number    | Exactly `4 << n` bytes corresponding to the IEEE 754 binary representation of the floating-point number | Preceded by 1 byte. 0 indicates not null.                    |
| Complex Number           | Exactly `4 << n` bytes for each floating-point number `a` and `b` where the complex number is `a + bi`. | Preceded by 1 byte. 0 indicates not null.                    |
| Enum                     | Stored as an unsigned fixed-size integer (see above), where `n` is determined by the number of enumerated values | Stored as **signed** fixed-size integer, and most significant bit of first byte indicates null if set. |
| Boolean                  | A single boolean value is encoded as 1 byte where 0 indicates `false` and any other value indicates `true`. | A single boolean value is encoded as 1 byte. The most significant bit indicates null if set. |
| Fixed-Length String      | UTF-8 encoding of the string, padded with spaces to fit within allotted space. | Preceded by 1 byte. 0 indicates not null.                    |
| Variable-Length String   | Length of the string is encoded as an unsigned variable-size integer followed by UTF-8 encoding of the string | Length of the string is encoded as a *signed* variable-size integer, and most significant bit indicates null if set. |
| Fixed-Length Array       | List of values encoded using the type specified by the schema | Preceded by 1 byte. 0 indicates not null.                    |
| Variable-Length Array    | Length of array encoded much like a string above. Arrays of boolean and/or nullable values may be optimized. | Length of the array is encoded as a *signed* variable-size integer, and most significant bit indicates null if set. |
| Object w/fixed fields    | Encoded values for each field in the order specified by the schema. Nullable fields may be optimized. | Preceded by 1 byte. 0 indicates not null.                    |
| Object w/variable fields | Number of entries encoded much like a string above. Key-value pairs are encoded using the types specified by the schema | Number of entries is encoded as a *signed* variable-size integer, and most significant bit indicates null if set. |
| Variant                  | Variable-size unsigned integer indicating the schema ID to use to decode this value. The actual value then follows. | Variants themselves are not nullable.                        |

[^3]: If the value is null, only 1 byte is written.

### Optimizations

The following optimizations modify the above rules for encoding values:

* For objects either nullable or boolean fields, a single bitmask is placed a the start of the object
* For arrays of a nullable type, each group of 8 elements is preceded by the corresponding null bitmask
* Arrays of type boolean are encoded as bitmasks.  The final byte's least significant bits are padded with zeros

## Schema JSON Specification

### Types

| Type                     | JSON Type Name | Additional Options                                           |
| ------------------------ | -------------- | ------------------------------------------------------------ |
| Fixed-size Integer       | int            | * `signed` - boolean indicating if integer is signed or unsigned<br />* `bits` - one of the following numbers indicating the size of the integer: 8, 16, 32, 64, 128, 256, 512, 1024 |
| Variable-size Integer    | int            | * `signed` - boolean indicating if integer is signed or unsigned<br />* `bits` - must be `null` or omitted |
| Floating-point Number    | float          | * `bits` - one of the following numbers indicating the size of the floating-point: 32, 64 |
| Complex Number           | complex        | * `bits` - one of the following numbers indicating the size of the complex number: 64, 128 |
| Enum                     | enum           | * `values` - an object mapping strings to integer values     |
| Boolean                  | bool           |                                                              |
| Fixed-Length String      | string         | * `length` - the length of the string in bytes               |
| Variable-Length String   | string         | * `length` - must be `null` or omitted                       |
| Fixed-Length Array       | array          | * `length` - the length of the string in bytes               |
| Variable-Length Array    | array          | * `length` - must be `null` or omitted                       |
| Object w/fixed fields    | object         | * `fields` - an array of fields. Each field is an type object with keys:<br />`name`[^4], `type`, and any additional options for the `type` |
| Object w/variable fields | object         | * `fields` - must be `null` or omitted                       |
| Variant                  | variant        |                                                              |

[^4]: It is strongly encouraged to use [camelCase](https://en.wikipedia.org/wiki/Camel_case) for object field names.

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
      "type": "integer",
      "signed": false,
      "size": 1
    }
  ]
}
```

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
// NewSchema decodes a schema stored in buf and returns an error if the schema is invalid
func NewSchema(buf []byte) (Schema, error)
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