package schemer

import (
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
)

type FloatSchema struct {
	Bits         int // must be 32 or 64
	WeakDecoding bool
	IsNullable   bool
}

func (s *FloatSchema) Valid() bool {
	return s.Bits == 32 || s.Bits == 64
}

// fixme
func (s *FloatSchema) MarshalJSON() ([]byte, error) {
	return nil, nil
}

// Bytes encodes the schema in a portable binary format
func (s *FloatSchema) Bytes() []byte {

	// floating point schemas are 1 byte long
	var schema []byte = make([]byte, 1)

	schema[0] = 0b01010000 // bit pattern for floating point number schema

	// The most signifiant bit indicates whether or not the type is nullable
	if s.IsNullable {
		schema[0] |= 1
	}

	// bit 2 unused

	// bit 3 = floating-point size in (32 << n) bits
	if s.Bits == 32 {
		// do nothing; third bit should be 0
	} else if s.Bits == 64 {
		// third bit should be one; indicating 64 bit floating point
		schema[0] |= 4
	}

	// bit 4 = is reserved

	return schema
}

// Encode uses the schema to write the encoded value of v to the output stream
func (s *FloatSchema) Encode(w io.Writer, i interface{}) error {

	if i == nil {
		return fmt.Errorf("cannot encode nil value. To encode a null, pass in a null pointer")
	}

	// just double check the schema they are using
	if !s.Valid() {
		return fmt.Errorf("cannot encode using invalid StringSchema schema")
	}

	v := reflect.ValueOf(i)

	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		v = v.Elem()
	}

	if s.IsNullable {
		// did the caller pass in a nil value, or a null pointer?
		if !v.IsValid() {

			fmt.Println("value encoded as a null...")

			// per the revised spec, 1 indicates null
			w.Write([]byte{1})
			return nil
		} else {
			// 0 indicates not null
			w.Write([]byte{0})
		}
	} else {
		// if nullable is false
		// but they are trying to encode a nil value.. then that is an error
		if !v.IsValid() {
			return fmt.Errorf("cannot enoded nil value when IsNullable is false")
		}
		// 0 indicates not null
		w.Write([]byte{0})
	}

	t := v.Type()
	k := t.Kind()

	if k != reflect.Float32 && k != reflect.Float64 {
		return fmt.Errorf("FloatSchema only supports encoding float32 and float64 values")
	}
	floatV := v.Float()

	// what type of value did they pass in?
	switch s.Bits {
	case 32:
		if k == reflect.Float64 {
			return fmt.Errorf("32bit FloatSchema schema cannot encode 64 bit values")
		}
		i := math.Float32bits(float32(floatV))

		n, err := w.Write([]byte{
			byte(i),
			byte(i >> 8),
			byte(i >> 16),
			byte(i >> 24),
		})
		if err == nil && n != 4 {
			err = errors.New("unexpected number of bytes written")
		}
		return err

	case 64:
		i := math.Float64bits(floatV)

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
		return err

	default:
		// error
	}
	return nil
}

func (s *FloatSchema) Decode(r io.Reader, i interface{}) error {

	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}

	return s.DecodeValue(r, reflect.ValueOf(i))
}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s *FloatSchema) DecodeValue(r io.Reader, v reflect.Value) error {

	// just double check the schema they are using
	if !s.Valid() {
		return fmt.Errorf("cannot decode using invalid floating point schema")
	}

	// first byte indicates whether value is null or not....
	buf := make([]byte, 1)
	_, err := io.ReadAtLeast(r, buf, 1)
	if err != nil {
		return err
	}
	valueIsNull := (buf[0] == 1)

	// if the data indicates this type is nullable, then the actual floating point
	// value is preceeded by one byte [which indicates if the encoder encoded a nill value]
	if s.IsNullable {
		if valueIsNull {
			if v.Kind() == reflect.Ptr {
				if v.CanSet() {
					v.Set(reflect.Zero(v.Type()))
					return nil
				}
				v = v.Elem()
				if v.CanSet() {
					v.Set(reflect.Zero(v.Type()))
					return nil
				}
				return fmt.Errorf("destination not settable")
			} else {
				return fmt.Errorf("cannot decode null value to non pointer to pointer type")
			}
		}
	}

	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		if v.IsNil() {
			if !v.CanSet() {
				return fmt.Errorf("decode destination is not settable")
			}
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	t := v.Type()
	k := t.Kind()

	var decodedFloat64 float64

	// take a look at the schema
	// the .Valid check above already ensured that Bits is either 32 or 64
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

	// Ensure v is settable
	if !v.CanSet() {
		return fmt.Errorf("decode destination is not settable")
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

		if decodedFloat64 == math.Trunc(decodedFloat64) {
			v.SetInt(int64(decodedFloat64))
		} else {
			return fmt.Errorf("loss of floating point precision not allowed w/o WeakDecoding")
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
		if decodedFloat64 < 0 {
			return fmt.Errorf("cannot decode negative ComplexNumber to unsigned int")
		}

		if decodedFloat64 == math.Trunc(decodedFloat64) {
			v.SetUint(uint64(decodedFloat64))
		} else {
			return fmt.Errorf("loss of floating point precision not allowed w/o WeakDecoding")
		}

	case reflect.Complex64:
		if !s.WeakDecoding {
			return fmt.Errorf("weak decoding not enabled; cannot decode float to Complex64")
		}
		v.SetComplex(complex(decodedFloat64, 0))

	case reflect.Complex128:
		if !s.WeakDecoding {
			return fmt.Errorf("weak decoding not enabled; cannot decode float to Complex128")
		}
		v.SetComplex(complex(decodedFloat64, 0))

	case reflect.String:
		if !s.WeakDecoding {
			return fmt.Errorf("weak decoding not enabled; cannot decode to string")
		}
		v.SetString(strconv.FormatFloat(decodedFloat64, 'f', -1, 64))

	default:
		return fmt.Errorf("invalid destination: %v", k)
	}

	return nil
}

func (s *FloatSchema) Nullable() bool {
	return s.IsNullable
}

func (s *FloatSchema) SetNullable(n bool) {
	s.IsNullable = n
}
