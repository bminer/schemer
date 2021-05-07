package schemer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

const maxFloatInt = int64(1)<<53 - 1
const minFloatInt = -maxFloatInt
const maxIntUint = uint64(1)<<63 - 1

type FixedIntSchema struct {
	Signed         bool
	Bits           int
	WeakDecoding   bool
	StrictEncoding bool
	IsNullable     bool
}

func (s FixedIntSchema) IsValid() bool {
	return s.Bits == 8 || s.Bits == 16 || s.Bits == 32 || s.Bits == 64
}

// Bytes encodes the schema in a portable binary format
func (s FixedIntSchema) Bytes() []byte {

	// fixed length schemas are 1 byte long total
	var schema []byte = make([]byte, 1)

	schema[0] = 0b00000000 // bit pattern for fixed int

	// The most signifiant bit indicates whether or not the type is nullable
	if s.IsNullable {
		schema[0] |= 1
	}

	// bit3 indicates if the the fixed length int is signed or not
	if s.Signed {
		schema[0] |= 4
	}

	//
	switch s.Bits {
	case 8:
		//do nothing
	case 16:
		schema[0] |= 8
	case 32:
		schema[0] |= 16
	case 64:
		schema[0] |= 24
	default:
	}

	return schema

}

// if this function is called MarshalJSON it seems to be called
// recursively by the json library???
func (s FixedIntSchema) DoMarshalJSON() ([]byte, error) {
	if !s.IsValid() {
		return nil, fmt.Errorf("invalid floating point schema")
	}

	return json.Marshal(s)
}

// if this function is called UnmarshalJSON it seems to be called
// recursively by the json library???
func (s FixedIntSchema) DoUnmarshalJSON(buf []byte) error {
	return json.Unmarshal(buf, s)
}

func writeUint(w io.Writer, v uint64, s FixedIntSchema) error {
	switch s.Bits {
	case 8:
		n, err := w.Write([]byte{byte(v)})
		if err == nil && n != 1 {
			err = errors.New("unexpected number of bytes written")
		}
		return err
	case 16:
		n, err := w.Write([]byte{
			byte(v),
			byte(v >> 8),
		})
		if err == nil && n != 2 {
			err = errors.New("unexpected number of bytes written")
		}
		return err
	case 32:
		n, err := w.Write([]byte{
			byte(v),
			byte(v >> 8),
			byte(v >> 16),
			byte(v >> 24),
		})
		if err == nil && n != 4 {
			err = errors.New("unexpected number of bytes written")
		}
		return err
	case 64:
		n, err := w.Write([]byte{
			byte(v),
			byte(v >> 8),
			byte(v >> 16),
			byte(v >> 24),
			byte(v >> 32),
			byte(v >> 40),
			byte(v >> 48),
			byte(v >> 56),
		})
		if err == nil && n != 8 {
			err = errors.New("unexpected number of bytes written")
		}
		return err
	case 128:
		n, err := w.Write([]byte{
			byte(v),
			byte(v >> 8),
			byte(v >> 16),
			byte(v >> 24),
			byte(v >> 32),
			byte(v >> 40),
			byte(v >> 48),
			byte(v >> 56),
			0, 0, 0, 0, 0, 0, 0, 0,
		})
		if err == nil && n != 16 {
			err = errors.New("unexpected number of bytes written")
		}
		return err
	default:
		return fmt.Errorf("invalid fixed integer size: %d bits", s.Bits)
	}
}

func readUint(r io.Reader, s FixedIntSchema) (uint64, error) {
	const errVal = uint64(0)

	// Read len(buf) bytes from r
	buf := make([]byte, int(s.Bits)/8)
	_, err := io.ReadAtLeast(r, buf, len(buf))
	if err != nil {
		return errVal, err
	}

	// Convert bytes to int value
	switch s.Bits {
	case 8:
		var raw int8 = int8(buf[0])
		return uint64(raw), nil
	case 16:
		var raw int16 = int16(buf[0]) |
			int16(buf[1])<<8
		return uint64(raw), nil
	case 32:
		var raw int32 = int32(buf[0]) |
			int32(buf[1])<<8 |
			int32(buf[2])<<16 |
			int32(buf[3])<<24
		return uint64(raw), nil
	case 64:
		return uint64(buf[0]) |
			uint64(buf[1])<<8 |
			uint64(buf[2])<<16 |
			uint64(buf[3])<<24 |
			uint64(buf[4])<<32 |
			uint64(buf[5])<<40 |
			uint64(buf[6])<<48 |
			uint64(buf[7])<<56, nil
	case 128:
		return uint64(buf[0]) |
			uint64(buf[1])<<8 |
			uint64(buf[2])<<16 |
			uint64(buf[3])<<24 |
			uint64(buf[4])<<32 |
			uint64(buf[5])<<40 |
			uint64(buf[6])<<48 |
			uint64(buf[7])<<56, nil
		// TODO: Check that remaining buf is all zeroes
	default:
		return errVal, fmt.Errorf("invalid fixed integer size: %d bits", s.Bits)
	}
}

// CheckType returns true if the integer type passed for i
// matched the schema
func checkType(s FixedIntSchema, k reflect.Kind) bool {

	var typeOK bool

	switch k {
	case reflect.Int:
		if uintSize == 32 {
			typeOK = s.Bits == 32 && s.Signed
		} else {
			typeOK = s.Bits == 64 && s.Signed
		}
	case reflect.Int8:
		typeOK = s.Bits == 8 && s.Signed
	case reflect.Int16:
		typeOK = s.Bits == 16 && s.Signed
	case reflect.Int32:
		typeOK = s.Bits == 32 && s.Signed
	case reflect.Int64:
		typeOK = s.Bits == 64 && s.Signed

	case reflect.Uint:
		if uintSize == 32 {
			typeOK = s.Bits == 32 && !s.Signed
		} else {
			typeOK = s.Bits == 64 && !s.Signed
		}
	case reflect.Uint8:
		typeOK = s.Bits == 8 && !s.Signed
	case reflect.Uint16:
		typeOK = s.Bits == 16 && !s.Signed
	case reflect.Uint32:
		typeOK = s.Bits == 32 && !s.Signed
	case reflect.Uint64:
		typeOK = s.Bits == 64 && !s.Signed
	default:
		typeOK = false
	}

	return typeOK
}

// Encode uses the schema to write the encoded value of v to the output stream
func (s FixedIntSchema) Encode(w io.Writer, i interface{}) error {

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot encode using invalid FixedIntSchema schema")
	}

	if s.IsNullable {
		// did the caller pass in a nil value, or a null pointer
		if i == nil ||
			(reflect.TypeOf(i).Kind() == reflect.Ptr && reflect.ValueOf(i).IsNil()) {
			// we encode a null value by writing a single non 0 byte
			w.Write([]byte{1})
			return nil
		} else {
			// 0 means not null (with actual encoded bytes to follow)
			w.Write([]byte{0})
		}
	} else {
		if i == nil {
			return fmt.Errorf("cannot enoded nil value when IsNullable is false")
		}
	}

	v := reflect.ValueOf(i)
	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		v = v.Elem()
	}
	t := v.Type()
	k := t.Kind()

	if !checkType(s, k) {
		return fmt.Errorf("encode failure; value to be encoded does not match FixedIntSchema schema")
	}

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
		end >>= (64 - s.Bits)
		if s.Signed {
			end /= 2
			start = -int64(end) - 1
		}
		if intVal > int64(end) || intVal < start {
			return fmt.Errorf("integer out of range %d to %d", start, end)
		}
		// Write value
		uintVal := uint64(intVal) << 1
		if intVal < 0 {
			uintVal = ^uintVal
		}
		return writeUint(w, uintVal, s)
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		uintVal := v.Uint()
		// Check integer range
		start := int64(0)
		end := uint64(0xFFFFFFFFFFFFFFFF) // 8 bytes
		end >>= (64 - s.Bits)
		if s.Signed {
			end /= 2
			start = -int64(end) - 1
		}
		if uintVal > end {
			return fmt.Errorf("integer out of range %d to %d", start, end)
		}
		// Write value
		return writeUint(w, uintVal, s)
	}
	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s FixedIntSchema) Decode(r io.Reader, i interface{}) error {

	v := reflect.ValueOf(i)

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot encode using invalid FixedIntSchema schema")
	}

	// if the schema indicates this type is nullable, then the actual floating point
	// value is preceeded by one byte [which indicates if the encoder encoded a nill value]
	if s.IsNullable {
		buf := make([]byte, 1)
		_, err := io.ReadAtLeast(r, buf, 1)
		if err != nil {
			return err
		}
		if buf[0] != 0 {
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
				if v.Kind() == reflect.Ptr {
					// special way to return a nil pointer
					v.Set(reflect.Zero(v.Type()))
				} else {
					return fmt.Errorf("cannot decode null value to non pointer to pointer type")
				}
			} else {
				return fmt.Errorf("cannot decode null value to non pointer to pointer type")
			}
			return nil
		}
	}

	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		v = v.Elem()
	}
	t := v.Type()
	k := t.Kind()

	// Ensure v is settable
	if !v.CanSet() {
		return fmt.Errorf("decode destination is not settable")
	}

	// Decode value
	if s.Signed {
		uintVal, err := readUint(r, s)
		if err != nil {
			return err
		}
		intVal := int64(uintVal >> 1)
		if uintVal&1 != 0 {
			intVal = ^intVal
		}
		// Write to destination
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
			if v.OverflowInt(intVal) {
				return fmt.Errorf("decoded value %d overflows destination %v", intVal, k)
			}
			v.SetInt(intVal)
		case reflect.Uint:
			fallthrough
		case reflect.Uint8:
			fallthrough
		case reflect.Uint16:
			fallthrough
		case reflect.Uint32:
			fallthrough
		case reflect.Uint64:
			if intVal < 0 {
				return fmt.Errorf("decoded value %d incompatible with %v", intVal, k)
			}
			uintVal := uint64(intVal)
			if v.OverflowUint(uintVal) {
				return fmt.Errorf("decoded value %d overflows destination %v", uintVal, k)
			}
			v.SetUint(uintVal)
		case reflect.Float32:
			fallthrough
		case reflect.Float64:
			if intVal > maxFloatInt || intVal < minFloatInt {
				return fmt.Errorf("decoded value %d overflows destination %v", intVal, k)
			}
			vFloat := float64(intVal)
			if v.OverflowFloat(vFloat) {
				return fmt.Errorf("decoded value %d overflows destination %v", intVal, k)
			}
			v.SetFloat(vFloat)
		case reflect.Complex64:
			fallthrough
		case reflect.Complex128:
			if intVal > maxFloatInt || intVal < minFloatInt {
				return fmt.Errorf("decoded value %d overflows destination %v", intVal, k)
			}
			vComplex := complex(float64(intVal), 0)
			if v.OverflowComplex(vComplex) {
				return fmt.Errorf("decoded value %d overflows destination %v", intVal, k)
			}
			v.SetComplex(vComplex)
		case reflect.Bool:
			if !s.WeakDecoding {
				return fmt.Errorf("decoded value %d incompatible with %v", intVal, k)
			}
			v.SetBool(intVal != 0)
		case reflect.String:
			if !s.WeakDecoding {
				return fmt.Errorf("decoded value %d incompatible with %v", intVal, k)
			}
			v.SetString(strconv.FormatInt(intVal, 10))
		default:
			return fmt.Errorf("decoded value %d incompatible with %v", intVal, k)
		}
	} else {
		// Unsigned
		uintVal, err := readUint(r, s)
		if err != nil {
			return err
		}
		// Write to destination
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
			if v.OverflowUint(uintVal) {
				return fmt.Errorf("decoded value %d overflows destination %v", uintVal, k)
			}
			v.SetUint(uintVal)
		case reflect.Uint:
			fallthrough
		case reflect.Uint8:
			fallthrough
		case reflect.Uint16:
			fallthrough
		case reflect.Uint32:
			fallthrough
		case reflect.Uint64:
			uintVal := uint64(uintVal)
			if v.OverflowUint(uintVal) {
				return fmt.Errorf("decoded value %d overflows destination %v", uintVal, k)
			}
			v.SetUint(uintVal)
		case reflect.Float32:
			fallthrough
		case reflect.Float64:
			if uintVal > uint64(maxFloatInt) {
				return fmt.Errorf("decoded value %d overflows destination %v", uintVal, k)
			}
			vFloat := float64(uintVal)
			if v.OverflowFloat(vFloat) {
				return fmt.Errorf("decoded value %d overflows destination %v", uintVal, k)
			}
			v.SetFloat(vFloat)
		case reflect.Complex64:
			fallthrough
		case reflect.Complex128:
			if uintVal > uint64(maxFloatInt) {
				return fmt.Errorf("decoded value %d overflows destination %v", uintVal, k)
			}
			vComplex := complex(float64(uintVal), 0)
			if v.OverflowComplex(vComplex) {
				return fmt.Errorf("decoded value %d overflows destination %v", uintVal, k)
			}
			v.SetComplex(vComplex)
		case reflect.Bool:
			if !s.WeakDecoding {
				return fmt.Errorf("decoded value %d incompatible with %v", uintVal, k)
			}
			v.SetBool(uintVal != 0)
		case reflect.String:
			if !s.WeakDecoding {
				return fmt.Errorf("decoded value %d incompatible with %v", uintVal, k)
			}
			v.SetString(strconv.FormatUint(uintVal, 10))
		default:
			return fmt.Errorf("decoded value %d incompatible with %v", uintVal, k)
		}
	}
	return nil
}
