package schemer

import (
	"fmt"
	"io"
	"reflect"
)

type Schema interface {
	// Encode uses the schema to write the encoded value of v to the output stream
	Encode(w io.Writer, v interface{}) error
	// Decode uses the schema to read the next encoded value from the input stream and store it in v
	Decode(r io.Reader, v interface{}) error
}

// basicSchema represents a simple basic type like a number or boolean
type basicSchema struct {
	Header byte
	// TypesAllowed []Schema // ???
}

func (s basicSchema) Encode(w io.Writer, i interface{}) error {
	hb := s.Header
	v := reflect.ValueOf(i)
	t := v.Type()
	k := t.Kind()

	nullable := hb&0x80 > 0
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
				return fmt.Errorf("integer out of range %d to %d", start, end)
			}
			// Write value
			var n int
			var err error
			switch sizeBytes {
			case 1:
				n, err = w.Write([]byte{byte(intVal)})
			case 2:
				n, err = w.Write([]byte{
					byte(intVal),
					byte(intVal >> 8),
				})
			case 4:
				n, err = w.Write([]byte{
					byte(intVal),
					byte(intVal >> 8),
					byte(intVal >> 16),
					byte(intVal >> 24),
				})
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
				return fmt.Errorf("integer out of range %d to %d", start, end)
			}
			// Write value
			var n int
			var err error
			switch sizeBytes {
			case 1:
				n, err = w.Write([]byte{byte(intVal)})
			case 2:
				n, err = w.Write([]byte{
					byte(intVal),
					byte(intVal >> 8),
				})
			case 4:
				n, err = w.Write([]byte{
					byte(intVal),
					byte(intVal >> 8),
					byte(intVal >> 16),
					byte(intVal >> 24),
				})
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
	return nil
}
