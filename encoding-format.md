# Schemer Encoding Format

This document outlines how schemer encodes schemata and values.

## Schema

Schemata are used to encode and decode values, but schemata themselves can be encoded into a binary format. This format is described below.

The first byte of a schema encodes the most important type information as follows:

- The most signifiant bit indicates whether or not the type is nullable
- The 6 least significant bits identify the type according to the table below

| Type                  | Least Significant Bits | Notes                                                        |
| --------------------- | ---------------------- | ------------------------------------------------------------ |
| Fixed-size Integer    | 0b00 nnns              | where s is the signed/unsigned bit and n represents the encoded integer size in (8 << n) bits. Example: 0x07 is a signed 64-bit integer. |
| Variable-size Integer | 0b01 000s              | where s is the signed/unsigned bit                           |
| Floating-point Number | 0b01 01*n              | where n is the floating-point size in (32 << n) bits and * is reserved for future use |
| Complex Number        | 0b01 10*n              | where n is the complex number size in (64 << n) bits and * is reserved for future use |
| Boolean               | 0b01 1100              |                                                              |
| Enum                  | 0b01 1101              |                                                              |
| String                | 0b10 000f              | where f indicates that the string is of fixed byte length    |
| Array                 | 0b10 010f              | where f indicates that the array is of fixed length          |
| Object                | 0b10 100f              | where f indicates that the object has fixed number of fields |
| Variant               | 0b10 1100              |                                                              |
| Schema                | 0b10 1101              |                                                              |
| Custom Type           | 0b11 1111              | next 16 bytes is a UUID for the custom type followed by the raw schema |

## Values

The following table describes how schemer encodes different values. Nullable values are generally preceded with an additional byte to indicate whether or not the value is null.

| Type                     | Encoding Format                                              | Nullable Format [^1]                                         |
| ------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
| Fixed-size Integer       | Exactly `1 >> n` bytes of two's complement integer representation in little-endian byte order | Preceded by 1 byte. 0 indicates not null.                    |
| Variable-size Integer    | Each byte's most significant bit indicates more bytes follow. Lower 7 bits are concatenated (in big-endian order) to form [ZigZag-encoded integer](https://developers.google.com/protocol-buffers/docs/encoding?csw=1#types). | Initial byte can only store 6-bit value, and most significant bit indicates null if set. |
| Floating-point Number    | Exactly `4 << n` bytes corresponding to the IEEE 754 binary representation of the floating-point number | Preceded by 1 byte. 0 indicates not null.                    |
| Complex Number           | Exactly `4 << n` bytes for each floating-point number `a` and `b` where the complex number is `a + bi`. | Preceded by 1 byte. 0 indicates not null.                    |
| Enum                     | Stored as an unsigned fixed-size integer (see above), where `n` is determined by the number of enumerated values | Stored as **signed** fixed-size integer, and sign bit indicates null if set. |
| Boolean                  | A single boolean value is encoded as 1 byte. 0 indicates `false` and any other value indicates `true`. | A single boolean value is encoded as 1 byte. The most significant bit indicates null if set. |
| Fixed-Length String      | UTF-8 encoding of the string, padded with spaces to fit within allotted space. | Preceded by 1 byte. 0 indicates not null.                    |
| Variable-Length String   | Length of the string is encoded as an unsigned variable-size integer followed by UTF-8 encoding of the string | Length of the string is encoded as a **signed** variable-size integer, and sign bit indicates null if set. |
| Fixed-Length Array       | List of values encoded using the type specified by the schema | Preceded by 1 byte. 0 indicates not null.                    |
| Variable-Length Array    | Length of array encoded much like a string above. Arrays of boolean and/or nullable values may be optimized. | Length of the array is encoded as a **signed** variable-size integer, and sign bit indicates null if set. |
| Object w/fixed fields    | Encoded values for each field in the order specified by the schema. | Preceded by 1 byte. 0 indicates not null. Many nullable fields may be optimized to fit snugly into a bit map. |
| Object w/variable fields | Number of entries encoded much like a string above. Key-value pairs are encoded using the types specified by the schema | Number of entries is encoded as a **signed** variable-size integer, and sign bit indicates null if set. |
| Schema                   | An encoded schemer schema. If the schema is larger than 17 bytes, a custom type schema UUID is usually written rather than the entire schema. | If null, the encoded schema will indicate a nullable schema. |
| Variant                  | Schema for the written value followed by the actual value. If the schema is larger than 17 bytes, a custom type schema UUID is usually written rather than the entire schema. | If null, the encoded schema will indicate a nullable variant. |

[^1]: If the value is null, only 1 byte is written.

### Optimizations

The following optimizations modify the above rules for encoding values:

* For objects with a fixed number of nullable or boolean fields, a single bit map is placed a the start of the object
* For arrays of a nullable type, each group of 8 elements is preceded by the corresponding null bit map
* Arrays of type boolean are encoded as bit maps. The final byte's least significant bits are padded with zeros.
* Variable-length arrays and objects w/variable fields may be stored in blocks, where each block indicates the number of elements or key-value pairs. A block size of 0 indicates the end of the array or object. A negative block size indicates that the block size is followed by the number of bytes in the block.