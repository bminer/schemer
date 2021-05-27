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
	SchemaOptions
	Signed bool
}

// Bytes encodes the schema in a portable binary format
func (s *VarIntSchema) Bytes() []byte {

	// fixed length schemas are 1 byte long total
	var schema []byte = []byte{0b00010000}

	// The most signifiant bit indicates whether or not the type is nullable
	if s.SchemaOptions.Nullable {
		schema[0] |= 128
	}

	// next bit indicates if the the fixed length int is signed or not
	if s.Signed {
		schema[0] |= 1
	}

	return schema

}

func (s *VarIntSchema) MarshalJSON() ([]byte, error) {

	tmpMap := make(map[string]interface{}, 2)
	tmpMap["type"] = "int"
	tmpMap["signed"] = strconv.FormatBool(s.Signed)
	tmpMap["nullable"] = strconv.FormatBool(s.SchemaOptions.Nullable)

	return json.Marshal(tmpMap)
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
func (s *VarIntSchema) Encode(w io.Writer, i interface{}) error {

	v := reflect.ValueOf(i)

	ok, err := PreEncode(s, w, &v)
	if err != nil {
		return err
	}
	if !ok {
		return nil
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

func (s *VarIntSchema) Decode(r io.Reader, i interface{}) error {

	if i == nil {
		return fmt.Errorf("cannot decode to nil destination")
	}

	v := reflect.ValueOf(i)

	return s.DecodeValue(r, v)
}

// Decode uses the schema to read the next encoded value from the input stream and store it in v
func (s *VarIntSchema) DecodeValue(r io.Reader, v reflect.Value) error {

	ok, err := PreDecode(s, r, &v)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	t := v.Type()
	k := t.Kind()

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
		// Ensure v is settable
		if !v.CanSet() {
			return fmt.Errorf("decode destination is not settable")
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

func (s *VarIntSchema) Nullable() bool {
	return s.SchemaOptions.Nullable
}

func (s *VarIntSchema) SetNullable(n bool) {
	s.SchemaOptions.Nullable = n
}
