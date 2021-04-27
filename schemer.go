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

func isPtrToSignedInt(v interface{}) bool {
	return checkforPtrToInt(v, reflect.Int8) ||
		checkforPtrToInt(v, reflect.Int16) ||
		checkforPtrToInt(v, reflect.Int32) ||
		checkforPtrToInt(v, reflect.Int64)
}

func isPtrToUnsignedInt(v interface{}) bool {
	return checkforPtrToInt(v, reflect.Uint8) ||
		checkforPtrToInt(v, reflect.Uint16) ||
		checkforPtrToInt(v, reflect.Uint32) ||
		checkforPtrToInt(v, reflect.Uint64)
}

func decodeSignedFixedSizeInt(sizeInBits int, r io.Reader, v interface{}) error {

	var va reflect.Value

	switch sizeInBits {
	case 8:
		// does the value that the caller of this routine passed in make sense?
		if isPtrToSignedInt(v) {
			buf := make([]byte, 1)
			r.Read(buf)

			var ValueToWrite int8 = int8(buf[0])

			switch reflect.ValueOf(v).Elem().Kind() {
			case reflect.Int8:
				// we can always fit an 8bit integer into an 8bit integer
				var tmp8 int8 = int8(ValueToWrite)
				va = reflect.ValueOf(tmp8)
				reflect.ValueOf(v).Elem().Set(va)
			case reflect.Int16:
				// we can always fit an 8bit integer into a 16bit integer
				var tmp16 int16 = int16(ValueToWrite)
				va = reflect.ValueOf(tmp16)
				reflect.ValueOf(v).Elem().Set(va)
			case reflect.Int32:
				// we can always fit a 16bit integer into a 32bit integer
				var tmp32 int32 = int32(ValueToWrite)
				va = reflect.ValueOf(tmp32)
				reflect.ValueOf(v).Elem().Set(va)
			case reflect.Int64:
				// we can always fit a 16bit integer into a 64bit integer
				var tmp64 int64 = int64(ValueToWrite)
				va = reflect.ValueOf(tmp64)
				reflect.ValueOf(v).Elem().Set(va)
			}

		} else {
			return fmt.Errorf("cannot decode signed int8 into passed in destination")
		}
	case 16:
		// does the value that the caller of this routine passed in make sense?
		if isPtrToSignedInt(v) {
			buf := make([]byte, 2)
			r.Read(buf)

			var ValueToWrite int16 = int16(buf[0]) |
				int16(buf[1])<<8

			switch reflect.ValueOf(v).Elem().Kind() {
			case reflect.Int8:
				if ValueToWrite >= -128 && ValueToWrite < 127 {
					var tmp8 int8 = int8(ValueToWrite)
					va = reflect.ValueOf(tmp8)
					reflect.ValueOf(v).Elem().Set(va)
				} else {
					return fmt.Errorf("decoded value cannot fit into destination")
				}
			case reflect.Int16:
				// we can always fit a 16bit integer into a 16bit integer
				var tmp16 int16 = int16(ValueToWrite)
				va = reflect.ValueOf(tmp16)
				reflect.ValueOf(v).Elem().Set(va)
			case reflect.Int32:
				// we can always fit a 16bit integer into a 32bit integer
				var tmp32 int32 = int32(ValueToWrite)
				va = reflect.ValueOf(tmp32)
				reflect.ValueOf(v).Elem().Set(va)
			case reflect.Int64:
				// we can always fit a 16bit integer into a 64bit integer
				var tmp64 int64 = int64(ValueToWrite)
				va = reflect.ValueOf(tmp64)
				reflect.ValueOf(v).Elem().Set(va)
			}
		} else {
			return fmt.Errorf("cannot decode signed int16 into passed in destination")
		}
	case 32:
		// does the value that the caller of this routine passed in make sense?
		if isPtrToSignedInt(v) {
			buf := make([]byte, 4)
			r.Read(buf)

			var ValueToWrite int32 = int32(buf[0]) |
				int32(buf[1])<<8 |
				int32(buf[2])<<16 |
				int32(buf[3])<<24

			switch reflect.ValueOf(v).Elem().Kind() {
			case reflect.Int8:
				if ValueToWrite >= -128 && ValueToWrite < 127 {
					var tmp8 int8 = int8(ValueToWrite)
					va = reflect.ValueOf(tmp8)
					reflect.ValueOf(v).Elem().Set(va)
				} else {
					return fmt.Errorf("decoded value cannot fit into destination")
				}
			case reflect.Int16:
				if ValueToWrite >= -32768 && ValueToWrite < 32767 {
					var tmp16 int16 = int16(ValueToWrite)
					va = reflect.ValueOf(tmp16)
					reflect.ValueOf(v).Elem().Set(va)
				} else {
					return fmt.Errorf("decoded value cannot fit into destination")
				}
			case reflect.Int32:
				// we can always fit a 32bit integer into a 32it integer
				var tmp32 int32 = int32(ValueToWrite)
				va = reflect.ValueOf(tmp32)
				reflect.ValueOf(v).Elem().Set(va)
			case reflect.Int64:
				// we can always fit a 32bit integer into a 64bit integer
				var tmp64 int64 = int64(ValueToWrite)
				va = reflect.ValueOf(tmp64)
				reflect.ValueOf(v).Elem().Set(va)
			}
		} else {
			return fmt.Errorf("cannot decode signed int32 into passed in destination")
		}
	case 64:
		// does the value that the caller of this routine passed in make sense?
		if isPtrToSignedInt(v) {
			buf := make([]byte, 8)
			r.Read(buf)

			var ValueToWrite int64 = int64(buf[0]) |
				int64(buf[1])<<8 |
				int64(buf[2])<<16 |
				int64(buf[3])<<24 |
				int64(buf[4])<<32 |
				int64(buf[5])<<40 |
				int64(buf[6])<<48 |
				int64(buf[7])<<56

			switch reflect.ValueOf(v).Elem().Kind() {
			case reflect.Int8:
				if ValueToWrite >= -128 && ValueToWrite < 127 {
					var tmp8 int8 = int8(ValueToWrite)
					va = reflect.ValueOf(tmp8)
					reflect.ValueOf(v).Elem().Set(va)
				} else {
					return fmt.Errorf("decoded value cannot fit into destination")
				}
			case reflect.Int16:
				if ValueToWrite >= -32768 && ValueToWrite < 32767 {
					var tmp16 int16 = int16(ValueToWrite)
					va = reflect.ValueOf(tmp16)
					reflect.ValueOf(v).Elem().Set(va)
				} else {
					return fmt.Errorf("decoded value cannot fit into destination")
				}
			case reflect.Int32:
				if ValueToWrite >= -2147483648 && ValueToWrite < 2147483647 {
					var tmp32 int32 = int32(ValueToWrite)
					va = reflect.ValueOf(tmp32)
					reflect.ValueOf(v).Elem().Set(va)
				} else {
					return fmt.Errorf("decoded value cannot fit into destination")
				}
			case reflect.Int64:
				// we can always fit a 64bit integer into a 64bit integer
				var tmp64 int64 = int64(ValueToWrite)
				va = reflect.ValueOf(tmp64)
				reflect.ValueOf(v).Elem().Set(va)
			}

		} else {
			return fmt.Errorf("cannot decode signed int64 into passed in destination")
		}
	}

	return nil
}

func decodeUnsignedFixedSizeInt(sizeInBits int, r io.Reader, v interface{}) error {

	var va reflect.Value

	switch sizeInBits {
	case 8:
		// does the value that the caller of this routine passed in make sense?
		if isPtrToUnsignedInt(v) {
			buf := make([]byte, 1)
			r.Read(buf)

			var ValueToWrite uint8 = uint8(buf[0])

			switch reflect.ValueOf(v).Elem().Kind() {
			case reflect.Uint8:
				// we can always fit an 8bit integer into an 8bit integer
				var tmp8 uint8 = uint8(ValueToWrite)
				va = reflect.ValueOf(tmp8)
				reflect.ValueOf(v).Elem().Set(va)
			case reflect.Uint16:
				// we can always fit an 8bit integer into a 16bit integer
				var tmp16 uint16 = uint16(ValueToWrite)
				va = reflect.ValueOf(tmp16)
				reflect.ValueOf(v).Elem().Set(va)
			case reflect.Uint32:
				// we can always fit a 16bit integer into a 32bit integer
				var tmp32 uint32 = uint32(ValueToWrite)
				va = reflect.ValueOf(tmp32)
				reflect.ValueOf(v).Elem().Set(va)
			case reflect.Uint64:
				// we can always fit a 16bit integer into a 64bit integer
				var tmp64 uint64 = uint64(ValueToWrite)
				va = reflect.ValueOf(tmp64)
				reflect.ValueOf(v).Elem().Set(va)
			}

		} else {
			return fmt.Errorf("cannot decode signed int8 into passed in destination")
		}

	case 64:
		// does the value that the caller of this routine passed in make sense?
		if isPtrToUnsignedInt(v) {
			buf := make([]byte, 8)
			r.Read(buf)

			var ValueToWrite uint64 = uint64(buf[0]) |
				uint64(buf[1])<<8 |
				uint64(buf[2])<<16 |
				uint64(buf[3])<<24 |
				uint64(buf[4])<<32 |
				uint64(buf[5])<<40 |
				uint64(buf[6])<<48 |
				uint64(buf[7])<<56

			switch reflect.ValueOf(v).Elem().Kind() {
			case reflect.Uint8:
				if ValueToWrite <= 255 {
					var tmp8 uint8 = uint8(ValueToWrite)
					va = reflect.ValueOf(tmp8)
					reflect.ValueOf(v).Elem().Set(va)
				} else {
					return fmt.Errorf("decoded value cannot fit into destination")
				}
			case reflect.Uint16:
				if ValueToWrite <= 65535 {
					var tmp16 uint16 = uint16(ValueToWrite)
					va = reflect.ValueOf(tmp16)
					reflect.ValueOf(v).Elem().Set(va)
				} else {
					return fmt.Errorf("decoded value cannot fit into destination")
				}
			case reflect.Uint32:
				if ValueToWrite < 2147483647 {
					var tmp32 uint32 = uint32(ValueToWrite)
					va = reflect.ValueOf(tmp32)
					reflect.ValueOf(v).Elem().Set(va)
				} else {
					return fmt.Errorf("decoded value cannot fit into destination")
				}
			case reflect.Uint64:
				// we can always fit a 64bit integer into a 64bit integer
				var tmp64 uint64 = uint64(ValueToWrite)
				va = reflect.ValueOf(tmp64)
				reflect.ValueOf(v).Elem().Set(va)
			}

		} else {
			return fmt.Errorf("cannot decode signed int64 into passed in destination")
		}

	}

	return nil
}

func decodeFixedSizeInt(hb byte, r io.Reader, v interface{}) error {

	// determine the size of the fixed-length integer
	signed, sizeInBits := sizeAndSignOfFixedLenInt(hb)

	if signed {
		return decodeSignedFixedSizeInt(sizeInBits, r, v)
	} else {
		return decodeUnsignedFixedSizeInt(sizeInBits, r, v)
	}

	return nil
}

func (s BasicSchema) Decode(r io.Reader, v interface{}) error {

	hb := s.Header

	if headerByteToConst(hb) == FixedSizeInteger_Type {
		return decodeFixedSizeInt(hb, r, v)
	} else {
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
