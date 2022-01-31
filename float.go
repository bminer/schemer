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

type FloatSchema struct {
	SchemaOptions
	Bits int // must be 32 or 64
}

func (s *FloatSchema) GoType() reflect.Type {
	var retval reflect.Type

	if s.Bits == 32 {
		var t float32
		retval = reflect.TypeOf(t)
	} else {
		var t float64
		retval = reflect.TypeOf(t)
	}

	if s.Nullable() {
		retval = reflect.PtrTo(retval)
	}

	return retval
}

func (s *FloatSchema) Valid() bool {
	return s.Bits == 32 || s.Bits == 64
}

func (s *FloatSchema) MarshalJSON() ([]byte, error) {
	if !s.Valid() {
		return nil, fmt.Errorf("invalid FloatSchema")
	}
	return json.Marshal(map[string]interface{}{
		"type":     "float",
		"nullable": s.Nullable(),
		"version":  SchemerVersion,
		"bits":     s.Bits,
	})
}

// Bytes encodes the schema in a portable binary format
func (s *FloatSchema) MarshalSchemer() ([]byte, error) {

	// floating point schemas are 1 byte long
	var schema []byte = []byte{FloatByte}

	// The most signifiant bit indicates whether or not the type is nullable
	if s.Nullable() {
		schema[0] |= NullMask
	}

	if s.Bits == 32 {
		// do nothing; bit 1 = 0
	} else if s.Bits == 64 {
		// set bit 1; indicating 64 bit floating point
		schema[0] |= 1
	}

	return schema, nil
}

// Encode uses the schema to write the encoded value of i to the output stream
func (s *FloatSchema) Encode(w io.Writer, i interface{}) error {
	return s.EncodeValue(w, reflect.ValueOf(i))
}

// EncodeValue uses the schema to write the encoded value of v to the output stream
func (s *FloatSchema) EncodeValue(w io.Writer, v reflect.Value) error {

	// just double check the schema they are using
	if !s.Valid() {
		return fmt.Errorf("cannot encode using invalid StringSchema schema")
	}

	done, err := PreEncode(w, &v, s.Nullable())
	if err != nil || done {
		return err
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

// Decode uses the schema to read the next encoded value from the input stream and store it in i
func (s *FloatSchema) Decode(r io.Reader, i interface{}) error {
	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}
	return s.DecodeValue(r, reflect.ValueOf(i))
}

// DecodeValue uses the schema to read the next encoded value from the input stream and store it in v
func (s *FloatSchema) DecodeValue(r io.Reader, v reflect.Value) error {

	// just double check the schema they are using
	if !s.Valid() {
		return fmt.Errorf("cannot decode using invalid floating point schema")
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
		if !s.WeakDecoding() {
			return fmt.Errorf("weak decoding not enabled; cannot decode float to Complex64")
		}
		v.SetComplex(complex(decodedFloat64, 0))

	case reflect.Complex128:
		if !s.WeakDecoding() {
			return fmt.Errorf("weak decoding not enabled; cannot decode float to Complex128")
		}
		v.SetComplex(complex(decodedFloat64, 0))

	case reflect.String:
		if !s.WeakDecoding() {
			return fmt.Errorf("weak decoding not enabled; cannot decode to string")
		}
		v.SetString(strconv.FormatFloat(decodedFloat64, 'f', -1, 64))

	default:
		return fmt.Errorf("invalid destination: %v", k)
	}

	return nil
}
