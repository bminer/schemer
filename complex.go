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

type ComplexSchema struct {
	SchemaOptions

	Bits int // must be 64 or 128
}

func (s *ComplexSchema) GoType() reflect.Type {
	var c1 complex64
	var c2 complex128
	var retval reflect.Type

	if s.Bits == 64 {
		retval = reflect.TypeOf(c1)
	} else {
		retval = reflect.TypeOf(c2)
	}

	if s.Nullable() {
		retval = reflect.PtrTo(retval)
	}

	return retval
}

func (s *ComplexSchema) Valid() bool {
	return s.Bits == 64 || s.Bits == 128
}

func (s *ComplexSchema) MarshalJSON() ([]byte, error) {
	if !s.Valid() {
		return nil, fmt.Errorf("invalid ComplexSchema")
	}
	return json.Marshal(map[string]interface{}{
		"version":  SchemerVersion,
		"type":     "complex",
		"nullable": s.Nullable(),
		"bits":     s.Bits,
	})
}

// Bytes encodes the schema in a portable binary format
func (s *ComplexSchema) MarshalSchemer() ([]byte, error) {

	// floating point schemas are 1 byte long
	var schema []byte = []byte{ComplexByte}

	// bit 8 indicates whether or not the type is nullable
	if s.Nullable() {
		schema[0] |= NullMask
	}

	// bit 1 = complex number size in (64 << n) bits
	if s.Bits == 64 {
		// do nothing; third bit should be 0
	} else if s.Bits == 128 {
		// third bit should be one; indicating 128 bit complex
		schema[0] |= 1
	}

	return schema, nil
}

// Encode uses the schema to write the encoded value of i to the output stream
func (s *ComplexSchema) Encode(w io.Writer, i interface{}) error {
	return s.EncodeValue(w, reflect.ValueOf(i))
}

// EncodeValue uses the schema to write the encoded value of v to the output stream
func (s *ComplexSchema) EncodeValue(w io.Writer, v reflect.Value) error {

	// just double check the schema they are using
	if !s.Valid() {
		return fmt.Errorf("cannot encode using invalid ComplexNumber schema")
	}

	done, err := PreEncode(w, &v, s.Nullable())
	if err != nil || done {
		return err
	}

	t := v.Type()
	k := t.Kind()

	if k != reflect.Complex64 && k != reflect.Complex128 {
		return fmt.Errorf("ComplexSchema only supports encoding Complex64 and Complex128 values")
	}
	complex := v.Complex()

	switch s.Bits {
	case 64:
		if k == reflect.Float64 {
			return fmt.Errorf("32bit FloatSchema schema cannot encode 64 bit values")
		}
		r := math.Float32bits(float32(real(complex)))
		imaginary := math.Float32bits(float32(imag(complex)))

		n, err := w.Write([]byte{
			byte(r),
			byte(r >> 8),
			byte(r >> 16),
			byte(r >> 24),
			byte(imaginary),
			byte(imaginary >> 8),
			byte(imaginary >> 16),
			byte(imaginary >> 24),
		})
		if err == nil && n != 8 {
			err = errors.New("unexpected number of bytes written")
		}
		return err

	case 128:
		r := math.Float64bits(real(complex))
		imaginary := math.Float64bits(imag(complex))

		n, err := w.Write([]byte{
			byte(r),
			byte(r >> 8),
			byte(r >> 16),
			byte(r >> 24),
			byte(r >> 32),
			byte(r >> 40),
			byte(r >> 48),
			byte(r >> 56),
			byte(imaginary),
			byte(imaginary >> 8),
			byte(imaginary >> 16),
			byte(imaginary >> 24),
			byte(imaginary >> 32),
			byte(imaginary >> 40),
			byte(imaginary >> 48),
			byte(imaginary >> 56),
		})
		if err == nil && n != 16 {
			err = errors.New("unexpected number of bytes written")
		}
		return err
	default:
		// error
	}
	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in i
func (s *ComplexSchema) Decode(r io.Reader, i interface{}) error {
	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}
	return s.DecodeValue(r, reflect.ValueOf(i))
}

// DecodeValue uses the schema to read the next encoded value from the input stream and store it in v
func (s *ComplexSchema) DecodeValue(r io.Reader, v reflect.Value) error {

	// just double check the schema they are using
	if !s.Valid() {
		return fmt.Errorf("cannot decode using invalid ComplexNumber schema")
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

	var realPart float64
	var imagPart float64

	// take a look at the schema..
	switch s.Bits {
	case 64:
		buf := make([]byte, 8)
		_, err := io.ReadAtLeast(r, buf, 8)
		if err != nil {
			return err
		}
		realPart = float64(math.Float32frombits(
			uint32(buf[0]) |
				uint32(buf[1])<<8 |
				uint32(buf[2])<<16 |
				uint32(buf[3])<<24))
		imagPart = float64(math.Float32frombits(
			uint32(buf[4]) |
				uint32(buf[5])<<8 |
				uint32(buf[6])<<16 |
				uint32(buf[7])<<24))
	case 128:
		buf := make([]byte, 16)
		_, err := io.ReadAtLeast(r, buf, 16)
		if err != nil {
			return err
		}
		realPart = math.Float64frombits(
			uint64(buf[0]) |
				uint64(buf[1])<<8 |
				uint64(buf[2])<<16 |
				uint64(buf[3])<<24 |
				uint64(buf[4])<<32 |
				uint64(buf[5])<<40 |
				uint64(buf[6])<<48 |
				uint64(buf[7])<<56)
		imagPart = math.Float64frombits(
			uint64(buf[8]) |
				uint64(buf[9])<<8 |
				uint64(buf[10])<<16 |
				uint64(buf[11])<<24 |
				uint64(buf[12])<<32 |
				uint64(buf[13])<<40 |
				uint64(buf[14])<<48 |
				uint64(buf[15])<<56)

	}
	var complexToWrite complex128 = complex(realPart, imagPart)

	// Ensure v is settable
	if !v.CanSet() {
		return fmt.Errorf("decode destination is not settable")
	}

	switch k {
	case reflect.Complex64:
		fallthrough
	case reflect.Complex128:
		if v.OverflowComplex(complexToWrite) {
			return fmt.Errorf("decoded complex overflows destination %v", k)
		}
		v.SetComplex(complexToWrite)

	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		// make sure there is no imaginary component
		if imagPart != 0 {
			return fmt.Errorf("cannot decode ComplexNumber to non complex type when imaginary component is present")
		}

		if v.OverflowFloat(realPart) {
			return fmt.Errorf("decoded value overflows destination %v", k)
		}
		v.SetFloat(realPart)

	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		// make sure there is no imaginary component
		if imagPart != 0 {
			return fmt.Errorf("cannot decode ComplexNumber to non complex type when imaginary component is present")
		}

		if v.OverflowInt(int64(realPart)) {
			return fmt.Errorf("decoded value overflows destination %v", k)
		}

		if realPart == math.Trunc(realPart) {
			v.SetInt(int64(realPart))
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
		// make sure there is no imaginary component
		if imagPart != 0 {
			return fmt.Errorf("cannot decode ComplexNumber to non complex type when imaginary component is present")
		}
		if v.OverflowUint(uint64(realPart)) {
			return fmt.Errorf("decoded value overflows destination %v", k)
		}
		if realPart < 0 {
			return fmt.Errorf("cannot decode negative ComplexNumber to unsigned int")
		}

		if realPart == math.Trunc(realPart) {
			v.SetUint(uint64(realPart))
		} else {
			return fmt.Errorf("loss of floating point precision not allowed w/o WeakDecoding")
		}

	case reflect.String:
		if !s.WeakDecoding() {
			return fmt.Errorf("weak decoding not enabled; cannot decode to string")
		}
		tmp := complex128(complex(realPart, imagPart))
		v.SetString(strconv.FormatComplex(tmp, 'E', -1, int(s.Bits)))

	case reflect.Slice:
		fallthrough
	case reflect.Array:
		if !s.WeakDecoding() {
			return fmt.Errorf("weak decoding not enabled; cannot decode complex to array/slice")
		}

		if v.Len() != 2 {
			return fmt.Errorf("complex numbers must be decoded into array/slice of exactly length 2")
		}

		elemK := t.Elem().Kind()
		if elemK != reflect.Float32 && elemK != reflect.Float64 {
			return fmt.Errorf("complex numbers must be decoded into array/slice of type float32 or float64")
		}

		// check overflow for each float

		v.Index(0).SetFloat(realPart)
		v.Index(1).SetFloat(imagPart)

	default:
		return fmt.Errorf("invalid destination %v", k)
	}

	return nil
}
