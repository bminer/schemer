package schemer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

/*
	- only schemer is versioned
	- every level of nested type has its own version included
	- composite types delete children's versions
	- extra byte prepended to each binary schema
	- no versioning or checksum or anything on binary data
*/

const uintSize = 32 << (^uint(0) >> 32 & 1) // 32 or 64

const maxFloatInt = int64(1)<<53 - 1
const minFloatInt = -maxFloatInt

type FixedIntSchema struct {
	SchemaOptions

	Signed bool
	Bits   int // must be 8, 16, 32, or 64
}

func (s *FixedIntSchema) GoType() reflect.Type {
	var retval reflect.Type

	if s.Signed {
		switch s.Bits {
		case 8:
			var t int8
			retval = reflect.TypeOf(t)
		case 16:
			var t int16
			retval = reflect.TypeOf(t)
		case 32:
			var t int32
			retval = reflect.TypeOf(t)
		case 64:
			var t int
			retval = reflect.TypeOf(t)
		}
	} else {
		switch s.Bits {
		case 8:
			var t uint8
			retval = reflect.TypeOf(t)
		case 16:
			var t uint16
			retval = reflect.TypeOf(t)
		case 32:
			var t uint32
			retval = reflect.TypeOf(t)
		case 64:
			var t uint
			retval = reflect.TypeOf(t)
		}
	}

	if s.Nullable() {
		retval = reflect.PtrTo(retval)
	}

	return retval
}

func (s *FixedIntSchema) Valid() bool {
	return s.Bits == 8 || s.Bits == 16 || s.Bits == 32 || s.Bits == 64
}

// Bytes encodes the schema in a portable binary format
func (s *FixedIntSchema) MarshalSchemer() ([]byte, error) {

	// FixedIntSchema is 1 byte long total + the schemer version
	var schema []byte = []byte{FixedIntByte, SchemerVersion}

	// bit8 indicates whether or not the type is nullable
	if s.Nullable() {
		schema[0] |= NullMask
	}

	// bit1 indicates if the the fixed length int is signed or not
	if s.Signed {
		schema[0] |= 1
	}

	//
	switch s.Bits {
	case 8:
		//do nothing
	case 16:
		schema[0] |= 2
	case 32:
		schema[0] |= 4
	case 64:
		schema[0] |= 6
	default:
	}

	return schema, nil

}

func (s *FixedIntSchema) MarshalJSON() ([]byte, error) {
	if !s.Valid() {
		return nil, fmt.Errorf("invalid FixedIntSchema schema")
	}
	return json.Marshal(map[string]interface{}{
		"version":  SchemerVersion,
		"type":     "int",
		"nullable": s.Nullable(),
		"bits":     s.Bits,
		"signed":   s.Signed,
	})
}

func writeUint(w io.Writer, v uint64, s *FixedIntSchema) error {
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
	default:
		return fmt.Errorf("invalid fixed integer size: %d bits", s.Bits)
	}
}

func readUint(r io.Reader, s *FixedIntSchema) (uint64, error) {
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
	default:
		return errVal, fmt.Errorf("invalid fixed integer size: %d bits", s.Bits)
	}
}

// checkType() returns true if reflect.Kind matches the passed in schema
// matched the schema
func checkType(s *FixedIntSchema, k reflect.Kind) bool {

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

// Encode uses the schema to write the encoded value of i to the output stream
func (s *FixedIntSchema) Encode(w io.Writer, i interface{}) error {
	return s.EncodeValue(w, reflect.ValueOf(i))
}

// EncodeValue uses the schema to write the encoded value of v to the output stream
func (s *FixedIntSchema) EncodeValue(w io.Writer, v reflect.Value) error {

	// just double check the schema they are using
	if !s.Valid() {
		return fmt.Errorf("cannot encode using invalid FixedIntSchema schema")
	}

	done, err := PreEncode(w, &v, s.Nullable())
	if err != nil || done {
		return err
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

// Decode uses the schema to read the next encoded value from the input stream and store it in i
func (s *FixedIntSchema) Decode(r io.Reader, i interface{}) error {
	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}
	return s.DecodeValue(r, reflect.ValueOf(i))
}

// DecodeValue uses the schema to read the next encoded value from the input stream and store it in v
func (s *FixedIntSchema) DecodeValue(r io.Reader, v reflect.Value) error {

	// just double check the schema they are using
	if !s.Valid() {
		return fmt.Errorf("cannot decode using invalid FixedIntSchema schema")
	}

	done, err := PreDecode(r, &v, s.Nullable())
	if err != nil || done {
		return err
	}

	t := v.Type()
	k := t.Kind()

	if k == reflect.Interface {
		v.Set(reflect.New(s.GoType()))

		v = v.Elem().Elem()
		t = v.Type()
		k = t.Kind()
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

		// Ensure v is settable
		if !v.CanSet() {
			return fmt.Errorf("decode destination is not settable")
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
			if !s.WeakDecoding() {
				return fmt.Errorf("decoded value %d incompatible with %v", intVal, k)
			}
			v.SetBool(intVal != 0)
		case reflect.String:
			if !s.WeakDecoding() {
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
			if v.OverflowInt(int64(uintVal)) {
				return fmt.Errorf("decoded value %d overflows destination %v", uintVal, k)
			}
			v.SetInt(int64(uintVal))
		case reflect.Uint:
			fallthrough
		case reflect.Uint8:
			fallthrough
		case reflect.Uint16:
			fallthrough
		case reflect.Uint32:
			fallthrough
		case reflect.Uint64:
			//uintVal := uint64(uintVal)
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
			if !s.WeakDecoding() {
				return fmt.Errorf("decoded value %d incompatible with %v", uintVal, k)
			}
			v.SetBool(uintVal != 0)
		case reflect.String:
			if !s.WeakDecoding() {
				return fmt.Errorf("decoded value %d incompatible with %v", uintVal, k)
			}
			v.SetString(strconv.FormatUint(uintVal, 10))
		default:
			return fmt.Errorf("decoded value %d incompatible with %v", uintVal, k)
		}
	}
	return nil
}
