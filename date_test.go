package schemer

import (
	"bytes"
	"testing"
	"time"
)

// uses a time.Time directly...
func testRegisteredType1(useJSON bool, t *testing.T) {

	var Date1 time.Time = time.Now()

	tmp, err := SchemaOf(&Date1)
	if err != nil {
		t.Error(err)
		return
	}

	writerSchema, ok := tmp.(*DateSchema)
	if !ok {
		t.Error("Incorrect (non DateSchema) type returned by SchemaOf")
		return
	}

	var encodedData bytes.Buffer

	err = writerSchema.Encode(&encodedData, Date1)
	if err != nil {
		t.Error(err)
		return
	}

	var binarywriterSchema []byte
	var readerSchema Schema

	if useJSON {
		// write out our schema as JSON
		binarywriterSchema, err = writerSchema.MarshalJSON()
		if err != nil {
			t.Error("writerSchema.MarshalJSON() failed", err)
		}

		// recreate the schema from the JSON
		readerSchema, err = DecodeSchemaJSON(bytes.NewReader(binarywriterSchema))
		if err != nil {
			t.Error("cannot create writerSchema from raw JSON data", err)
			return
		}
	} else {
		// write out our schema as binary
		binarywriterSchema, err = writerSchema.MarshalSchemer()
		if err != nil {
			t.Error("writerSchema.MarshalSchemer() failed", err)
		}

		// recreate the schema from the binary data
		readerSchema, err = DecodeSchema(binarywriterSchema)
		if err != nil {
			t.Error("cannot create writerSchema from raw binary data", err)
			return
		}
	}

	var Date2 time.Time

	r := bytes.NewReader(encodedData.Bytes())
	err = readerSchema.Decode(r, &Date2)
	if err != nil {
		t.Error(err)
		return
	}

	// convert original date.Now() to have millisecond precision
	origDate := time.Unix(0, (Date1.UnixNano()/1000000)*1000000)

	// and now make sure that the structs match!
	decodeOK := origDate.Equal(Date2)

	if !decodeOK {
		t.Error("unexpected custom data type (date) decode...")
	}

}

// uses no struct
func testRegisteredType2(useJSON bool, t *testing.T) {

	type SourceStruct struct {
		IntField1 int
		Date      time.Time
		Str       string
	}

	var structToEncode = SourceStruct{IntField1: 42, Date: time.Now(), Str: "test"}

	tmp, err := SchemaOf(&structToEncode)
	if err != nil {
		t.Error(err)
		return
	}

	writerSchema, ok := tmp.(*FixedObjectSchema)
	if !ok {
		t.Error("Incorrect (non DateSchema) type returned by SchemaOf")
		return
	}

	var encodedData bytes.Buffer

	err = writerSchema.Encode(&encodedData, structToEncode)
	if err != nil {
		t.Error(err)
		return
	}

	var binarywriterSchema []byte
	var readerSchema Schema

	if useJSON {
		// write out our schema as JSON
		binarywriterSchema, err = writerSchema.MarshalJSON()
		if err != nil {
			t.Error("writerSchema.MarshalJSON() failed", err)
		}

		// recreate the schema from the JSON
		readerSchema, err = DecodeSchemaJSON(bytes.NewReader(binarywriterSchema))
		if err != nil {
			t.Error("cannot create writerSchema from raw JSON data", err)
			return
		}
	} else {
		// write out our schema as binary
		binarywriterSchema, err = writerSchema.MarshalSchemer()
		if err != nil {
			t.Error("writerSchema.MarshalSchemer() failed", err)
		}

		// recreate the schema from the binary data
		readerSchema, err = DecodeSchema(binarywriterSchema)
		if err != nil {
			t.Error("cannot create writerSchema from raw binary data", err)
			return
		}
	}

	type DestinationStruct struct {
		IntField1 int
		Date      time.Time
		Str       string
	}

	var structToDecode = DestinationStruct{}

	r := bytes.NewReader(encodedData.Bytes())
	err = readerSchema.Decode(r, &structToDecode)
	if err != nil {
		t.Error(err)
		return
	}

	// convert original date.Now() to have millisecond precision
	origDate := time.Unix(0, (structToDecode.Date.UnixNano()/1000000)*1000000)

	// and now make sure that the structs match!
	decodeOK := true
	decodeOK = decodeOK && (structToDecode.IntField1 == structToEncode.IntField1)
	decodeOK = decodeOK && (structToDecode.Str == structToEncode.Str)
	decodeOK = decodeOK && (origDate.Equal(structToDecode.Date))

	if !decodeOK {
		t.Error("unexpected struct to struct decode, using custom data type (date)")
	}

}

func TestCustomDates(t *testing.T) {
	testRegisteredType1(true, t)
	testRegisteredType1(false, t)

	testRegisteredType2(true, t)
	testRegisteredType2(false, t)
}
