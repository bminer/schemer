package schemer

import (
	"testing"
)

func TestDecodeEnum(t *testing.T) {

	/*

		type Weekday int

		const (
			Sunday    Weekday = 0
			Monday    Weekday = 1
			Tuesday   Weekday = 2
			Wednesday Weekday = 3
			Thursday  Weekday = 4
			Friday    Weekday = 5
			Saturday  Weekday = 6
		)

		var myDay Weekday = Sunday

		/*
			fixedIntSchema := FixedInteger{Signed: true, Bits: 64, WeakDecoding: true}

			var buf bytes.Buffer
			var err error
			var valueToEncode int64 = 100
			buf.Reset()

			fmt.Println("Testing decoding int64 value")

			fixedIntSchema.Encode(&buf, myDay)
			if err != nil {
				t.Error(err)
			}

			// decode into signed integer types

			r := bytes.NewReader(buf.Bytes())

			fmt.Println("int64 to int")
			var decodedValue1 int
			err = fixedIntSchema.Decode(r, &decodedValue1)
			if err != nil {
				t.Error(err)
			}
	*/

}
