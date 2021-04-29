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

type ComplexNumberSchema struct {
	Bits         uint8 // must be 64 or 128
	WeakDecoding bool
}

func (s ComplexNumberSchema) IsValid() bool {
	return s.Bits == 64 || s.Bits == 128
}

func (s ComplexNumberSchema) MarshalJSON() ([]byte, error) {

	return json.Marshal(s)

}

func (s *ComplexNumberSchema) UnmarshalJSON(buf []byte) error {

	return json.Unmarshal(buf, s)

}

// Encode uses the schema to write the encoded value of v to the output stream
func (s ComplexNumberSchema) Encode(w io.Writer, v interface{}) error {
	value := reflect.ValueOf(v)
	t := value.Type()
	k := t.Kind()
	complex := value.Complex()

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot encode using invalid ComplexNumber schema")
	}

	switch k {
	case reflect.Complex64:

		real := math.Float32bits(float32(real(complex)))
		imaginary := math.Float32bits(float32(imag(complex)))

		n, err := w.Write([]byte{
			byte(real),
			byte(real >> 8),
			byte(real >> 16),
			byte(real >> 24),
			byte(imaginary),
			byte(imaginary >> 8),
			byte(imaginary >> 16),
			byte(imaginary >> 24),
		})
		if err == nil && n != 8 {
			return errors.New("unexpected number of bytes written")
		}

	case reflect.Complex128:
		real := math.Float64bits(real(complex))
		imaginary := math.Float64bits(imag(complex))

		n, err := w.Write([]byte{
			byte(real),
			byte(real >> 8),
			byte(real >> 16),
			byte(real >> 24),
			byte(real >> 32),
			byte(real >> 40),
			byte(real >> 48),
			byte(real >> 56),
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
			return errors.New("unexpected number of bytes written")
		}
	}
	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s ComplexNumberSchema) Decode(r io.Reader, i interface{}) error {
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
		return fmt.Errorf("cannot decode using invalid ComplexNumber schema")
	}

	var realPart float64
	var imaginaryPart float64

	// take a look at the schema..
	switch s.Bits {
	case 64:
		buf := make([]byte, 8)
		_, err := io.ReadAtLeast(r, buf, 8)
		if err != nil {
			return err
		}
		realPart = float64(math.Float32frombits(uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24))
		imaginaryPart = float64(math.Float32frombits(uint32(buf[4]) | uint32(buf[5])<<8 | uint32(buf[6])<<16 | uint32(buf[7])<<24))
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
		imaginaryPart = math.Float64frombits(
			uint64(buf[8]) |
				uint64(buf[9])<<8 |
				uint64(buf[10])<<16 |
				uint64(buf[11])<<24 |
				uint64(buf[12])<<32 |
				uint64(buf[13])<<40 |
				uint64(buf[14])<<48 |
				uint64(buf[15])<<56)

	}
	var complexToWrite complex128 = complex128(complex(realPart, imaginaryPart))

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
		if imaginaryPart != 0 {
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
		if imaginaryPart != 0 {
			return fmt.Errorf("cannot decode ComplexNumber to non complex type when imaginary component is present")
		}

		if v.OverflowInt(int64(realPart)) {
			return fmt.Errorf("decoded value overflows destination %v", k)
		}
		if s.WeakDecoding {
			// with weak decoding, we will allow loss of decimal point values
			v.SetInt(int64(realPart))
		} else {
			if realPart == math.Trunc(realPart) {
				v.SetInt(int64(realPart))
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
		// make sure there is no imaginary component
		if imaginaryPart != 0 {
			return fmt.Errorf("cannot decode ComplexNumber to non complex type when imaginary component is present")
		}
		if v.OverflowUint(uint64(realPart)) {
			return fmt.Errorf("decoded value overflows destination %v", k)
		}
		if realPart < 0 {
			return fmt.Errorf("cannot decode negative ComplexNumber to unsigned int")
		}
		if s.WeakDecoding {
			// with weak decoding, we will allow loss of decimal point values
			v.SetUint(uint64(realPart))
		} else {
			if realPart == math.Trunc(realPart) {
				v.SetUint(uint64(realPart))
			} else {
				return fmt.Errorf("loss of floating point precision not allowed w/o WeakDecoding")
			}
		}

	case reflect.String:
		if !s.WeakDecoding {
			return fmt.Errorf("weak decoding not enabled; cannot decode to string")
		}
		tmp := complex128(complex(realPart, imaginaryPart))
		v.SetString(strconv.FormatComplex(tmp, 'E', -1, int(s.Bits)))

	default:
		return fmt.Errorf("invalid destination %v", k)
	}

	return nil
}
