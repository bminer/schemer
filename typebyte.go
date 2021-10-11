package schemer

// Collection of bit masks and values for the type byte of an encoded schema
const (
	NullMask     = 0x80 // nullable bit
	CustomMask   = 0x40 // custom schema bit
	CustomIDMask = 0x3F // custom schema identifier

	// FixedInt is 0b000 nnns where s is the signed/unsigned bit and
	// n represents the encoded integer size in (8 << n) bits.
	FixedIntMask     = 0x70
	FixedIntByte     = 0x00
	FixedIntBitsMask = 0x0E // Note: Needs to be shifted right 1
	IntSignedMask    = 0x01

	VarIntMask = 0x7E // 0b001 000s where s is the signed/unsigned bit
	VarIntByte = 0x10

	// Float is 0b001 01nn where n is the floating-point size in (32 << n) bits
	FloatMask     = 0x7C
	FloatBitsMask = 0x03
	FloatByte     = 0x14

	// Complex is 0b001 10nn where n is the complex number size in (64 << n) bits
	ComplexMask     = 0x7C
	ComplexBitsMask = 0x03
	ComplexByte     = 0x18

	BoolMask = 0x7F // 0b001 1100
	BoolByte = 0x1C

	EnumMask = 0x7F // 0b001 1101
	EnumByte = 0x1D

	StringMask      = 0x7F // 0b010 000f where f indicates fixed-length string
	VarStringByte   = 0x20
	FixedStringByte = 0x21

	ArrayMask      = 0x7F // 0b010 010f where f indicates fixed-length array
	VarArrayByte   = 0x24
	FixedArrayByte = 0x25

	// Object is 0b010 100f where f indicates that the object has fixed fields
	ObjectMask      = 0x7F
	VarObjectByte   = 0x28
	FixedObjectByte = 0x29
)
