package schemer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
)

type ComplexNumber struct {
	Bits         uint8 // must be 64 or 128
	WeakDecoding bool
}

func (s ComplexNumber) IsValid() bool {
	return s.Bits == 32 || s.Bits == 64
}

func (s ComplexNumber) MarshalJSON() ([]byte, error) {

	return json.Marshal(s)

}

func (s *ComplexNumber) UnmarshalJSON(buf []byte) error {

	return json.Unmarshal(buf, s)

}

// Encode uses the schema to write the encoded value of v to the output stream
func (s ComplexNumber) Encode(w io.Writer, v interface{}) error {
	value := reflect.ValueOf(v)
	t := value.Type()
	k := t.Kind()
	complex := value.Complex()

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
func (s ComplexNumber) Decode(r io.Reader, i interface{}) error {
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

	var rPart32 float32
	var iPart32 float32

	var rPart64 float64
	var iPart64 float64

	// take a look at the schema..
	switch s.Bits {
	case 64:
		buf := make([]byte, 8)
		_, err := io.ReadAtLeast(r, buf, 8)
		if err != nil {
			return err
		}
		rPart32 = math.Float32frombits(uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24)
		iPart32 = math.Float32frombits(uint32(buf[4]) | uint32(buf[5])<<8 | uint32(buf[6])<<16 | uint32(buf[7])<<24)
	case 128:
		buf := make([]byte, 16)
		_, err := io.ReadAtLeast(r, buf, 16)
		if err != nil {
			return err
		}
		rPart64 = math.Float64frombits(
			uint64(buf[0]) |
				uint64(buf[1])<<8 |
				uint64(buf[2])<<16 |
				uint64(buf[3])<<24 |
				uint64(buf[4])<<32 |
				uint64(buf[5])<<40 |
				uint64(buf[6])<<48 |
				uint64(buf[7])<<56)
		iPart64 = math.Float64frombits(
			uint64(buf[8]) |
				uint64(buf[9])<<8 |
				uint64(buf[10])<<16 |
				uint64(buf[11])<<24 |
				uint64(buf[12])<<32 |
				uint64(buf[13])<<40 |
				uint64(buf[14])<<48 |
				uint64(buf[15])<<56)

	}

	// Write to destination
	switch k {
	case reflect.Complex64:
		v.SetComplex(complex128(complex(rPart32, iPart32)))

	case reflect.Complex128:
		v.SetComplex(complex128(complex(rPart64, iPart64)))

	case reflect.String:
		if !s.WeakDecoding {
			return fmt.Errorf("weak decoding not enabled; cannot decode to string")
		}
		//v.SetString(strconv.FormatComplex(complex128(complex(r, i))), 'E', -1, s.Bits)

	default:
		return fmt.Errorf("invalid destination %v", k)
	}

	return nil
}
