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
	Encode(w io.Writer, v interface{}) error
	// Decode uses the schema to read the next encoded value from the input stream and store it in v
	Decode(r io.Reader, v interface{}) error
}

// basicSchema represents a simple basic type like a number or boolean
type BasicSchema struct {
	Header byte
	// TypesAllowed []Schema // ???
}

func ArchitectureIs64Bits() bool {
	return uintSize == 64
}

// FixedSizeInt returns a schema byte representing a fixed sized integer.
// If i isn't a fixed type integer, we return 0 and an error
func (s BasicSchema) FixedSizeInt(i interface{}) (byte, error) {

	v := reflect.ValueOf(i)
	t := v.Type()
	k := t.Kind()

	switch k {
	// int can either be 32 or 64 bits wide, depending on the architecture
	case reflect.Int:
		if ArchitectureIs64Bits() {
			return 7, nil
		} else {
			return 5, nil
		}

	case reflect.Int8:
		return 1, nil
	case reflect.Int16:
		return 3, nil
	case reflect.Int32:
		return 5, nil
	case reflect.Int64:
		return 7, nil

	// int can either be 32 or 64 bits wide, depending on the architecture
	case reflect.Uint:
		if ArchitectureIs64Bits() {
			return 6, nil
		} else {
			return 4, nil
		}
	case reflect.Uint8:
		return 0, nil
	case reflect.Uint16:
		return 2, nil
	case reflect.Uint32:
		return 4, nil
	case reflect.Uint64:
		return 6, nil
	}

	return 0, fmt.Errorf("invalid integer type")

}

// Encode will use the schema to encode the passed in value. Upon success, return the number of bytes used
// to encode the value, which is written into the passed in io.Writer.
func (s BasicSchema) Encode(w io.Writer, i interface{}) (int, error) {
	hb := s.Header
	v := reflect.ValueOf(i)
	t := v.Type()
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
			intVal := v.Int()
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
				if err != nil {
					return n, err
				}
				return 1, nil
			case 2:
				n, err = w.Write([]byte{
					byte(intVal),
					byte(intVal >> 8),
				})
				if n != 2 {
					err = errors.New("unexpected number of bytes written")
				}
				if err != nil {
					return n, err
				}
				return 2, nil
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
				if err != nil {
					return n, err
				}
				return 4, nil
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
				if err != nil {
					return n, err
				}
				return 8, nil
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
				if err != nil {
					return n, err
				}
				return 16, nil
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
			intVal := v.Uint()
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
				if err != nil {
					return n, err
				}
				return 1, nil
			case 2:
				n, err = w.Write([]byte{
					byte(intVal),
					byte(intVal >> 8),
				})
				if n != 2 {
					err = errors.New("unexpected number of bytes written")
				}
				if err != nil {
					return n, err
				}
				return 2, nil
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
				if err != nil {
					return n, err
				}
				return 4, nil
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
				if err != nil {
					return n, err
				}
				return 8, nil
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
				if err != nil {
					return n, err
				}
				return 16, nil
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
