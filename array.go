package schemer

import (
	"encoding/json"
	"io"
)

type ArraySchema struct {
	WeakDecoding bool
}

func (s ArraySchema) IsValid() bool {
	return true
}

func (s ArraySchema) MarshalJSON() ([]byte, error) {

	return json.Marshal(s)

}

func (s *ArraySchema) UnmarshalJSON(buf []byte) error {

	return json.Unmarshal(buf, s)

}

// Encode uses the schema to write the encoded value of v to the output stream
func (s ArraySchema) Encode(w io.Writer, v interface{}) error {

	/*
		value := reflect.ValueOf(v)
		t := value.Type()
		k := t.Kind()
		complex := value.Complex()

		// just double check the schema they are using
		if !s.IsValid() {
			return fmt.Errorf("cannot encode using invalid ComplexNumber schema")
		}

		// take a look at what they are passing in...
		// make sure it is a slice or an array
		switch k {
		case reflect.Slice:
			fallthrough
		case reflect.Array:

			// this will tell us the type of the array they are passing in...
			v.Index(0).Kind()

			// figure out the length of the array they are talking about
			for i = 0; i < v.Len(); i++

		default:
			return errors.New("invalid encoding source; only arrays or slices can be encoded by ")

	*/

	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s ArraySchema) Decode(r io.Reader, i interface{}) error {

	/*
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

		case reflect.Slice:
			fallthrough
		case reflect.Array:
			if !s.WeakDecoding {
				return fmt.Errorf("weak decoding not enabled; cannot decode complex to array/slice")
			}

			if v.Len() != 2 {
				return fmt.Errorf("complex numbers must be decoded into array/slice of exactly length 2")
			}

			arrayOK := true
			arrayOK = arrayOK && (v.Index(0).Kind() == reflect.Float32 || v.Index(0).Kind() == reflect.Float64)
			arrayOK = arrayOK && (v.Index(1).Kind() == reflect.Float32 || v.Index(1).Kind() == reflect.Float64)

			if !arrayOK {
				return fmt.Errorf("complex numbers must be decoded into array/slice of type Float32 or Float64")
			}

			v.Index(0).SetFloat(realPart)
			v.Index(1).SetFloat(imaginaryPart)

		default:
			return fmt.Errorf("invalid destination %v", k)
		}
	*/

	return nil
}
