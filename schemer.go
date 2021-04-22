package schemer

import (
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

// basicSchema represents a simple basic type like a number or boolean
type BasicSchema struct {
	Header byte
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
	} else if hb&0xF0 == 0 {
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

func (s BasicSchema) Decode(r io.Reader, v interface{}) error {

	hb := s.Header

	// does the schema indicate a fixed length integer??
	if hb&0xF0 == 0 {

		// determine the size of the fixed-length integer
		tmp := (hb & 14) >> 1
		sizeInBits := 8 << tmp
		signed := tmp&0x01 > 0

		if signed {
			switch sizeInBits {
			//case 8:
			//case 16:
			//case 32:
			case 64:
				// does the value that the caller of this routine passed in make sense?
				if checkforPtrToInt(v, reflect.Int64) {

					buf := make([]byte, 8)
					lr := io.LimitReader(r, 8)
					bytesRead, _ := lr.Read(buf)

					fmt.Printf("total bytes read: %d \n", bytesRead)
					for i := 0; i < bytesRead; i++ {
						fmt.Printf("byte%d: %d\n", i, buf[i])
					}

					var ValueToWrite int64 = int64(buf[0]) |
						int64(buf[1])<<8 |
						int64(buf[2])<<16 |
						int64(buf[3])<<24 |
						int64(buf[4])<<32 |
						int64(buf[5])<<40 |
						int64(buf[6])<<48 |
						int64(buf[7])<<56

					va := reflect.ValueOf(ValueToWrite)
					reflect.ValueOf(v).Elem().Set(va)
				} else {
					return fmt.Errorf("cannot decode signed int64 into passed in passed in destination")
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

func (s BasicSchema) MarshalJSON() ([]byte, error) {
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
