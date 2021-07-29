package schemer

import (
	"bytes"
	"net"
	"testing"
)

// uses a net.IP  directly...
func testIPv41(useJSON bool, t *testing.T) {

	var srcIP net.IP = net.IPv4(192, 168, 0, 2)

	writerSchema := SchemaOf(&srcIP)

	var encodedData bytes.Buffer

	err := writerSchema.Encode(&encodedData, srcIP)
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
		readerSchema, err = DecodeJSONSchema(binarywriterSchema)
		if err != nil {
			t.Error("cannot create writerSchema from raw JSON data", err)
			return
		}
	} else {
		// write out our schema as binary
		binarywriterSchema = writerSchema.MarshalSchemer()
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

	var destIP net.IP

	r := bytes.NewReader(encodedData.Bytes())
	err = readerSchema.Decode(r, &destIP)
	if err != nil {
		t.Error(err)
		return
	}

	decodeOK := srcIP[0] == destIP[0] && srcIP[1] == destIP[1] && srcIP[2] == destIP[2] && srcIP[3] == destIP[3]

	if !decodeOK {
		t.Error("unexpected custom data type (IP) decode...")
	}

}

// uses a struct
func testIPv42(useJSON bool, t *testing.T) {

	type SourceStruct struct {
		IntField1 int
		IP        net.IP
		Str       string
	}

	var structToEncode = SourceStruct{IntField1: 42, IP: net.IPv4(192, 168, 0, 1), Str: "test"}

	writerSchema := SchemaOf(&structToEncode)

	var encodedData bytes.Buffer

	err := writerSchema.Encode(&encodedData, structToEncode)
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
		readerSchema, err = DecodeJSONSchema(binarywriterSchema)
		if err != nil {
			t.Error("cannot create writerSchema from raw JSON data", err)
			return
		}
	} else {
		// write out our schema as binary
		binarywriterSchema = writerSchema.MarshalSchemer()
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
		IP        net.IP
		Str       string
	}

	var structToDecode = DestinationStruct{}

	r := bytes.NewReader(encodedData.Bytes())
	err = readerSchema.Decode(r, &structToDecode)
	if err != nil {
		t.Error(err)
		return
	}

	// and now make sure that the structs match!
	decodeOK := true
	decodeOK = decodeOK && (structToDecode.IntField1 == structToEncode.IntField1)
	decodeOK = decodeOK && (structToDecode.Str == structToEncode.Str)
	decodeOK = decodeOK &&
		structToEncode.IP[0] == structToDecode.IP[0] &&
		structToEncode.IP[1] == structToDecode.IP[1] &&
		structToEncode.IP[2] == structToDecode.IP[2] &&
		structToEncode.IP[3] == structToDecode.IP[3]

	if !decodeOK {
		t.Error("unexpected struct to struct decode, using custom data type (net.IP)")
	}

}

func TestIPv41(t *testing.T) {
	testIPv41(true, t)
	testIPv41(false, t)

	testIPv42(true, t)
	testIPv42(false, t)
}
