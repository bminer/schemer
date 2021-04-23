package schemer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

// https://golangbyexample.com/go-size-range-int-uint/
const uintSize = 32 << (^uint(0) >> 32 & 1) // 32 or 64

type Schema interface {
	// Encode uses the schema to write the encoded value of v to the output stream
	Encode(w io.Writer, v interface{}) (int, error)
	// Decode uses the schema to read the next encoded value from the input stream and store it in v
	Decode(r io.Reader, v interface{}) error
	// Bytes encodes the schema in a portable binary format
	Bytes() []byte
	// String returns the schema in a human-readable format
	// String() string
	// MarshalJSON returns the JSON encoding of the schema
	MarshalJSON() ([]byte, error)
	// UnmarshalJSON updates the schema by decoding the JSON-encoded schema in buf
	UnmarshalJSON(buf []byte) error
	// Nullable returns true if and only if the type is nullable
	Nullable() bool
	// SetNullable sets the nullable flag for the schema
	SetNullable(n bool)
}

// types that we can support encoding
const (
	FixedSizeInteger_Type = iota
	VariableSizeInteger_Type
	FloatingPointNumber_Type
	ComplexNumber_Type
	Boolean_Type
	Enum_Type
	FixedLengthString_Type
	VariableLengthString_Type
	FixedLengthArray_Type
	VariableLengthArray_Type
	ObjectWithfixedFields_Type
	ObjectWithvariableFields_Type
)

// basicSchema represents a simple basic type like a number or boolean
type BasicSchema struct {
	Header byte
}

// headerByteToConst function needs to be filled in!
func headerByteToConst(headerByte byte) int {

	if headerByte&0xF0 == 0 {
		return FixedSizeInteger_Type
	}

	// fixme
	return 0

}

// Encode uses the schema to write the encoded value of v to the output stream
func (s BasicSchema) Encode(w io.Writer, v interface{}) (int, error) {
	hb := s.Header
	value := reflect.ValueOf(v)
	t := value.Type()
	k := t.Kind()

	//nullable := hb&0x80 > 0
	if hb&0x20 > 0 {
		// string, array, object, or variant
	} else if headerByteToConst(hb) == FixedSizeInteger_Type {
		// fixed-size integer
		tmp := hb & 0x0F
		signed := tmp&0x01 > 0
		tmp >>= 1
		// `tmp` is now `n`
		sizeBytes := 1 << tmp
		// Encode the value
		/*if weak {
			switch k {
			case reflect.Bool:

			case reflect.Float32:

			case reflect.Float64:

			case reflect.Complex64:

			case reflect.Complex128:

			case reflect.String:
			}
			return nil
		}*/
		switch k {
		case reflect.Int:
			fallthrough
		case reflect.Int8:
			fallthrough
		case reflect.Int16:
			fallthrough
		case reflect.Int32:
			fallthrough
		case reflect.Int64:
			intVal := value.Int()
			// Check integer range
			start := int64(0)
			end := uint64(0xFFFFFFFFFFFFFFFF) // 8 bytes
			end >>= 8 * (8 - sizeBytes)
			if signed {
				end /= 2
				start = -int64(end) - 1
			}
			if intVal > int64(end) || intVal < start {
				return 0, fmt.Errorf("integer out of range %d to %d", start, end)
			}
			// Write value
			var n int
			var err error
			switch sizeBytes {
			case 1:
				n, err = w.Write([]byte{byte(intVal)})
				if n != 1 {
					err = errors.New("unexpected number of bytes written")
				}
				return n, err
			case 2:
				n, err = w.Write([]byte{
					byte(intVal),
					byte(intVal >> 8),
				})
				if n != 2 {
					err = errors.New("unexpected number of bytes written")
				}
				return n, err
			case 4:
				n, err = w.Write([]byte{
					byte(intVal),
					byte(intVal >> 8),
					byte(intVal >> 16),
					byte(intVal >> 24),
				})
				if n != 4 {
					err = errors.New("unexpected number of bytes written")
				}
				return n, err
			case 8:
				n, err = w.Write([]byte{
					byte(intVal),
					byte(intVal >> 8),
					byte(intVal >> 16),
					byte(intVal >> 24),
					byte(intVal >> 32),
					byte(intVal >> 40),
					byte(intVal >> 48),
					byte(intVal >> 56),
				})
				if n != 8 {
					err = errors.New("unexpected number of bytes written")
				}
				return n, err
			case 16:
				n, err = w.Write([]byte{
					byte(intVal),
					byte(intVal >> 8),
					byte(intVal >> 16),
					byte(intVal >> 24),
					byte(intVal >> 32),
					byte(intVal >> 40),
					byte(intVal >> 48),
					byte(intVal >> 56),
					0, 0, 0, 0, 0, 0, 0, 0,
				})
				if n != 16 {
					err = errors.New("unexpected number of bytes written")
				}
				return n, err
			}
		case reflect.Uint:
			fallthrough
		case reflect.Uint8:
			fallthrough
		case reflect.Uint16:
			fallthrough
		case reflect.Uint32:
			fallthrough
		case reflect.Uint64:
			intVal := value.Uint()
			// Check integer range
			start := int64(0)
			end := uint64(0xFFFFFFFFFFFFFFFF) // 8 bytes
			end >>= 8 * (8 - sizeBytes)
			if signed {
				end /= 2
				start = -int64(end) - 1
			}
			if intVal > end {
				return 0, fmt.Errorf("integer out of range %d to %d", start, end)
			}
			// Write value
			var n int
			var err error
			switch sizeBytes {
			case 1:
				n, err = w.Write([]byte{byte(intVal)})
				if n != 1 {
					err = errors.New("unexpected number of bytes written")
				}
				return n, err
			case 2:
				n, err = w.Write([]byte{
					byte(intVal),
					byte(intVal >> 8),
				})
				if n != 2 {
					err = errors.New("unexpected number of bytes written")
				}
				return n, err
			case 4:
				n, err = w.Write([]byte{
					byte(intVal),
					byte(intVal >> 8),
					byte(intVal >> 16),
					byte(intVal >> 24),
				})
				if n != 4 {
					err = errors.New("unexpected number of bytes written")
				}
				return n, err
			case 8:
				n, err = w.Write([]byte{
					byte(intVal),
					byte(intVal >> 8),
					byte(intVal >> 16),
					byte(intVal >> 24),
					byte(intVal >> 32),
					byte(intVal >> 40),
					byte(intVal >> 48),
					byte(intVal >> 56),
				})
				if n != 8 {
					err = errors.New("unexpected number of bytes written")
				}
				return n, err
			case 16:
				n, err = w.Write([]byte{
					byte(intVal),
					byte(intVal >> 8),
					byte(intVal >> 16),
					byte(intVal >> 24),
					byte(intVal >> 32),
					byte(intVal >> 40),
					byte(intVal >> 48),
					byte(intVal >> 56),
					0, 0, 0, 0, 0, 0, 0, 0,
				})
				if n != 16 {
					err = errors.New("unexpected number of bytes written")
				}
				return n, err
			}
		}
	} else if hb&0xF8 == 0 {
		// variable-size integer or floating-point
		if hb&0xFE == 0x10 {
			// variable-size integer
		} else {
			// floating-point
		}
	} else {
		// ....
	}
	return 0, nil
}

// checkforPtrToInt returns true if v is a pointer to an integer type
func checkforPtrToInt(v interface{}, intType reflect.Kind) bool {

	value := reflect.ValueOf(v)

	// dereference the pointer
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
		// and then make sure the type of the thing pointed to is the expected type
		if value.Kind() == intType {
			return true
		}
	}

	return false
}

func sizeAndSignOfFixedLenInt(headerByte byte) (bool, int) {

	tmp := (headerByte & 14) >> 1
	sizeInBits := 8 << tmp
	signed := headerByte&0x01 > 0

	return signed, sizeInBits
}

func (s BasicSchema) Decode(r io.Reader, v interface{}) error {

	var va reflect.Value
	hb := s.Header

	if headerByteToConst(hb) == FixedSizeInteger_Type {

		// determine the size of the fixed-length integer
		signed, sizeInBits := sizeAndSignOfFixedLenInt(hb)

		if signed {
			switch sizeInBits {
			case 8:
				// does the value that the caller of this routine passed in make sense?
				if checkforPtrToInt(v, reflect.Int8) {

					buf := make([]byte, 1)
					lr := io.LimitReader(r, 1)
					lr.Read(buf)

					var ValueToWrite int8 = int8(buf[0])

					va = reflect.ValueOf(ValueToWrite)
					reflect.ValueOf(v).Elem().Set(va)
				} else {
					return fmt.Errorf("cannot decode signed int8 into passed in destination")
				}
			case 16:
				// does the value that the caller of this routine passed in make sense?
				if checkforPtrToInt(v, reflect.Int16) {

					buf := make([]byte, 2)
					lr := io.LimitReader(r, 2)
					lr.Read(buf)

					var ValueToWrite int16 = int16(buf[0]) |
						int16(buf[1])<<8

					va = reflect.ValueOf(ValueToWrite)
					reflect.ValueOf(v).Elem().Set(va)
				} else {
					return fmt.Errorf("cannot decode signed int16 into passed in destination")
				}
			case 32:
				// does the value that the caller of this routine passed in make sense?
				if checkforPtrToInt(v, reflect.Int32) {

					buf := make([]byte, 4)
					lr := io.LimitReader(r, 4)
					lr.Read(buf)

					var ValueToWrite int32 = int32(buf[0]) |
						int32(buf[1])<<8 |
						int32(buf[2])<<16 |
						int32(buf[3])<<24

					va = reflect.ValueOf(ValueToWrite)
					reflect.ValueOf(v).Elem().Set(va)
				} else {
					return fmt.Errorf("cannot decode signed int32 into passed in destination")
				}
			case 64:
				// does the value that the caller of this routine passed in make sense?
				if checkforPtrToInt(v, reflect.Int64) {

					buf := make([]byte, 8)
					lr := io.LimitReader(r, 8)
					lr.Read(buf)

					var ValueToWrite int64 = int64(buf[0]) |
						int64(buf[1])<<8 |
						int64(buf[2])<<16 |
						int64(buf[3])<<24 |
						int64(buf[4])<<32 |
						int64(buf[5])<<40 |
						int64(buf[6])<<48 |
						int64(buf[7])<<56

					va = reflect.ValueOf(ValueToWrite)
					reflect.ValueOf(v).Elem().Set(va)
				} else {
					return fmt.Errorf("cannot decode signed int64 into passed in destination")
				}
			}
		} else {
			/*
				switch sizeInBits {
				case 8:
				case 16:
				case 32:
				case 64:
				}
			*/
		}

	}

	return nil
}

func (s BasicSchema) Bytes() []byte {
	return nil
}

func jsonForFixedSizeInt(header byte) ([]byte, error) {

	type fixedIntInfo struct {
		TypeName string
		Signed   bool
		Bits     int
	}

	// determine the size of the fixed-length integer
	signed, sizeInBits := sizeAndSignOfFixedLenInt(header)

	m := fixedIntInfo{"int", signed, sizeInBits}
	return json.Marshal(m)

}

func (s BasicSchema) MarshalJSON() ([]byte, error) {

	if headerByteToConst(s.Header) == FixedSizeInteger_Type {
		return jsonForFixedSizeInt(s.Header)
	}

	return nil, nil
}

func (s BasicSchema) UnmarshalJSON(buf []byte) error {
	return nil
}

func (s BasicSchema) Nullable() bool {
	return false
}

func (s BasicSchema) SetNullable(n bool) {

}

// SchemaOfType returns a schema for the passed in goType. Note that integers are by default always
// encoded using variable length encoding
func SchemaOfType(t reflect.Type) Schema {

	k := t.Kind()

	switch k {
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		return VarIntegerSchema(true)
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		return VarIntegerSchema(false)
	}
	return nil
}

func SchemaOf(i interface{}) Schema {

	t := reflect.TypeOf(i)

	return SchemaOfType(t)

}

//Question: what should return type be here???
//func FixedIntegerSchema(signed bool, bits int) BasicSchema {

func FixedIntegerSchema(signed bool, bits int) Schema {

	var bs BasicSchema

	if signed {

		switch bits {
		case 8:
			bs.Header = 1
		case 16:
			bs.Header = 3
		case 32:
			bs.Header = 5
		case 64:
			bs.Header = 7
		default:
			return nil
		}
	} else {
		switch bits {
		case 8:
			bs.Header = 0
		case 16:
			bs.Header = 2
		case 32:
			bs.Header = 4
		case 64:
			bs.Header = 6
		default:
			return nil
		}

	}

	return bs

}

// VarIntegerSchema returns a basic schema for
func VarIntegerSchema(signed bool) Schema {

	var bs BasicSchema

	bs.Header = 16

	if signed {
		bs.Header |= 1
	}

	return bs
}

func ArchitectureIs64Bits() bool {
	return uintSize == 64
}
