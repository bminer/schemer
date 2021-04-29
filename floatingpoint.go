package schemer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
)

type FloatingPointNumber struct {
	Bits         uint8 // must be 64 or 128
	WeakDecoding bool
}

func (s FloatingPointNumber) IsValid() bool {
	return s.Bits == 32 || s.Bits == 64
}

// if this function is called MarshalJSON it seems to be called
// recursively by the json library???
func (s FloatingPointNumber) DoMarshalJSON() ([]byte, error) {

	if !s.IsValid() {
		return nil, fmt.Errorf("invalid floating point schema")
	}

	return json.Marshal(s)

}

// if this function is called UnmarshalJSON it seems to be called
// recursively by the json library???
func (s *FloatingPointNumber) DoUnmarshalJSON(buf []byte) error {

	return json.Unmarshal(buf, s)

}

// TODO: fixme
func Nullable() bool {
	return false
}

// TODO: fixme
func SetNullable(n bool) {
}

// Encode uses the schema to write the encoded value of v to the output stream
func (s FloatingPointNumber) Encode(w io.Writer, v interface{}) error {
	value := reflect.ValueOf(v)
	t := value.Type()
	k := t.Kind()
	floatToWrite := value.Float()

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot encode using invalid floating point schema")
	}

	// what type of value did they pass in?
	switch k {
	case reflect.Float32:
		i := math.Float32bits(float32(floatToWrite))

		n, err := w.Write([]byte{
			byte(i),
			byte(i >> 8),
			byte(i >> 16),
			byte(i >> 24),
		})
		if err == nil && n != 4 {
			return errors.New("unexpected number of bytes written")
		}
		if err != nil {
			return err
		}

	case reflect.Float64:
		i := math.Float64bits(floatToWrite)

		n, err := w.Write([]byte{
			byte(i),
			byte(i >> 8),
			byte(i >> 16),
			byte(i >> 24),
			byte(i >> 32),
			byte(i >> 40),
			byte(i >> 48),
			byte(i >> 56),
		})
		if err == nil && n != 8 {
			return errors.New("unexpected number of bytes written")
		}
		if err != nil {
			return err
		}

	}
	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s FloatingPointNumber) Decode(r io.Reader, i interface{}) error {
	v := reflect.ValueOf(i)
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

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot decode using invalid floating point schema")
	}

	var decodedFloat64 float64

	// take a look at the schema
	// the .IsValid check above already ensured that Bits is either 32 or 64
	switch s.Bits {
	case 32:
		buf := make([]byte, 4)
		_, err := io.ReadAtLeast(r, buf, 4)
		if err != nil {
			return err
		}
		var raw32 uint32 = uint32(buf[0]) |
			uint32(buf[1])<<8 |
			uint32(buf[2])<<16 |
			uint32(buf[3])<<24
		decodedFloat64 = float64(math.Float32frombits(raw32))
	case 64:
		buf := make([]byte, 8)
		_, err := io.ReadAtLeast(r, buf, 8)
		if err != nil {
			return err
		}

		var raw64 uint64 = uint64(buf[0]) |
			uint64(buf[1])<<8 |
			uint64(buf[2])<<16 |
			uint64(buf[3])<<24 |
			uint64(buf[4])<<32 |
			uint64(buf[5])<<40 |
			uint64(buf[6])<<48 |
			uint64(buf[7])<<56
		decodedFloat64 = math.Float64frombits(raw64)

	}

	// Write to destination
	switch k {

	case reflect.Float32:
		if v.OverflowFloat(decodedFloat64) {
			return fmt.Errorf("decoded value %f overflows destination %v", decodedFloat64, k)
		}
		v.SetFloat(decodedFloat64)

	case reflect.Float64:
		// we should always be able to decode a float64 into a float64...
		v.SetFloat(decodedFloat64)

	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		if v.OverflowInt(int64(decodedFloat64)) {
			return fmt.Errorf("decoded value %f overflows destination %v", decodedFloat64, k)
		}
		if s.WeakDecoding {
			v.SetInt(int64(decodedFloat64))
		} else {
			if decodedFloat64 == math.Trunc(decodedFloat64) {
				v.SetInt(int64(decodedFloat64))
			} else {
				return fmt.Errorf("loss of floating point precision not allowed w/o WeakDecoding")
			}
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
		if v.OverflowUint(uint64(decodedFloat64)) {
			return fmt.Errorf("decoded value %f overflows destination %v", decodedFloat64, k)
		}
		if s.WeakDecoding {
			v.SetUint(uint64(decodedFloat64))
		} else {
			if decodedFloat64 == math.Trunc(decodedFloat64) {
				v.SetUint(uint64(decodedFloat64))
			} else {
				return fmt.Errorf("loss of floating point precision not allowed w/o WeakDecoding")
			}
		}

	case reflect.Complex64:
		if !s.WeakDecoding {
			return fmt.Errorf("weak decoding not enabled; cannot decode float to Complex64")
		}
		v.SetComplex(complex128(complex(decodedFloat64, 0)))

	case reflect.Complex128:
		if !s.WeakDecoding {
			return fmt.Errorf("weak decoding not enabled; cannot decode float to Complex128")
		}
		v.SetComplex(complex128(complex(decodedFloat64, 0)))

	case reflect.String:
		if !s.WeakDecoding {
			return fmt.Errorf("weak decoding not enabled; cannot decode to string")
		}
		v.SetString(strconv.FormatFloat(decodedFloat64, 'f', -1, 64))

	default:
		return fmt.Errorf("invalid destination %v", k)
	}

	return nil
}
