TODO: remove this file

# Notes

- Done
  - JSON encoding / decoding
  - Nullable() / SetNullable()
  - Remove Create...Schema() functions
  - SchemaOfValue()
    - should be SchemaOfType() and only use type information
    - If a type is a pointer type (i.e. it got dereferenced), set IsNullable; otherwise,
      do not
    - Parsing of struct tags
      - Default tag to "schemer" but add option (to SchemaOf?) to customize?
      - Tag format: `schemer:"fieldName,extra_options"`
        or maybe `schemer:"[fieldName1,fieldName2]"`
        or to skip a field `schemer:"-"`
      - To discuss later...
  - DecodeSchema([]byte)
    - Only allocate Schema type when needed
    - Complete schema generation for nested structures
  - Bytes()
    - Return more than 1 byte, encode all schema info according to spec
  - Rename fixedlenstring to fixedstring
  - Remove spaces at top / bottom of functions / cases in switch statements, etc.
  - casing of variables
- 2020-05-25
  - schemer.go
    - `DecodeSchema` renamed to `DecodeSchema`
    - StructFieldOptions
      - Remove completely
      - objectField could become:
        - Aliases []string
        - Schema
    - Add `SchemaOptions` struct, which is inherited by all other Schema structs
      - `WeakDecoding bool`
      - `Nullable bool`
      - `IsNullable() bool`
      - `SetNullable(b bool)`
    - Add `DecodeSchemaJSON`
    - Rename `IsValid() bool` to `Valid() bool`
    - Remove `Valid() bool` where no validation is needed (i.e. `BoolSchema`)
    - Fix binary format:
      - Nullable bit is MSB (i.e. first bit from left to right) - 0x80 or `1 << 7` bitmask
      - 7th bit is unused
      - 6 least significant bits (i.e. bitmask 0x3F) are listed in table in documentation
  - bool.go
    - If `IsNullable()` (or probably better to use `.Nullable`), you should be able to encode `nil` I think
  - Encoding / Decoding
    - Helper function to handle "common" tasks before encoding / decoding
    - Encoding
      - Dereference `v`, the `reflect.Value` until:
        - `nil` is encountered _or_
        - `v.Kind()` is not a `Ptr` or `Interface`
      - If `s.Nullable`, write 1 if `v` is `nil`; otherwise, write 0
        - Continue...
      - Else If `v` is `nil`, return an error and abort
      - Else continue...
    - Decoding
      - If `s.Nullable`
        - If decoded value is null
          - If `v.Kind()` is `Ptr` or `Interface`, write `nil` to `v`
          - Else, return an error
      - Dereference `v` by continually initializing it recursively until `v.Kind()` is not `Ptr` or `Interface`
      - Continue...
