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
    - NewSchema([]byte)
        - Only allocate Schema type when needed
        - Complete schema generation for nested structures
    - Bytes()
        - Return more than 1 byte, encode all schema info according to spec
    - Rename fixedlenstring to fixedstring
    - Remove spaces at top / bottom of functions / cases in switch statements, etc.
    - casing of variables
- 2020-05-25
    - schemer.go
        - 