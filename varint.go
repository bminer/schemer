package schemer

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

type VarIntSchema struct {
	Signed       bool
	WeakDecoding bool
	IsNullable   bool
}

func (s VarIntSchema) IsValid() bool {
	return true
}

// Bytes encodes the schema in a portable binary format
func (s VarIntSchema) Bytes() []byte {

	// fixed length schemas are 1 byte long total
	var schema []byte = make([]byte, 1)

	schema[0] = 0b01000000 // bit pattern for fixed int

	// The most signifiant bit indicates whether or not the type is nullable
	if s.IsNullable {
		schema[0] |= 1
	}

	// next bit indicates if the the fixed length int is signed or not
	if s.Signed {
		schema[0] |= 4
	}

	return schema

}

// if this function is called MarshalJSON it seems to be called
// recursively by the json library???
func (s VarIntSchema) DoMarshalJSON() ([]byte, error) {
	if !s.IsValid() {
		return nil, fmt.Errorf("invalid floating point schema")
	}

	return json.Marshal(s)
}

// if this function is called UnmarshalJSON it seems to be called
// recursively by the json library???
func (s VarIntSchema) DoUnmarshalJSON(buf []byte) error {
	return json.Unmarshal(buf, s)
}

func writeVarUint(w io.Writer, v uint64) error {

	buf := make([]byte, binary.MaxVarintLen64)
	varIntBytes := binary.PutUvarint(buf, v)
	n, err := w.Write(buf[0:varIntBytes])
	if err == nil && n != varIntBytes {
		err = errors.New("unexpected number of bytes written")
	}

	return err
}

func readVarUint(r io.Reader) (uint64, error) {

	buf := make([]byte, binary.MaxVarintLen64)
	counter := 0

	// read one byte at a a time
	for {
		b := make([]byte, 1)
		_, err := io.ReadAtLeast(r, b, 1)
		if err != nil {
			return 0, err
		}
		buf[counter] = b[0]

		// keep reading out bytes until
		if b[0]&128 != 128 {
			break
		}

		counter++
	}

	decodedUInt, n := binary.Uvarint(buf)
	if n != counter+1 {
		return 0, fmt.Errorf("uvarint did not consume expected number of bytes")
	}

	return decodedUInt, nil

}

// Encode uses the schema to write the encoded value of v to the output stream
func (s VarIntSchema) Encode(w io.Writer, i interface{}) error {

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot encode using invalid VarIntSchema schema")
	}

	if i == nil {
		return fmt.Errorf("cannot encode nil value. To encode a null, pass in a null pointer")
	}

	if s.IsNullable {
		if reflect.TypeOf(i).Kind() == reflect.Ptr ||
			reflect.TypeOf(i).Kind() == reflect.Interface &&
				reflect.ValueOf(i).IsNil() {
			// we encode a null value by writing a single non 0 byte
			w.Write([]byte{1})
			return nil
		} else {
			// 0 means not null (with actual encoded bytes to follow)
			w.Write([]byte{0})
		}
	} else {
		if reflect.TypeOf(i).Kind() == reflect.Ptr ||
			reflect.TypeOf(i).Kind() == reflect.Interface &&
				reflect.ValueOf(i).IsNil() {
			return fmt.Errorf("cannot encode nil value when IsNullable is false")
		}
	}

	v := reflect.ValueOf(i)
	// Dereference pointer / interface types
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		v = v.Elem()
	}
	t := v.Type()
	k := t.Kind()

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
		// Write value
		uintVal := uint64(intVal) << 1
		if intVal < 0 {
			uintVal = ^uintVal
		}
		return writeVarUint(w, uintVal)
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
		return writeVarUint(w, uintVal)
	}

	return nil
}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s VarIntSchema) Decode(r io.Reader, i interface{}) error {

	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}

	v := reflect.ValueOf(i)

	// just double check the schema they are using
	if !s.IsValid() {
		return fmt.Errorf("cannot encode using invalid VarIntSchema schema")
	}

	// if the schema indicates this type is nullable, then the actual var int
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
		uintVal, err := readVarUint(r)
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
		uintVal, err := readVarUint(r)
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
