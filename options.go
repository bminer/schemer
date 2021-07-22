package schemer

// SchemaOptions are options common to each Schema
type SchemaOptions struct {
	nullable     bool
	weakDecoding bool
}

// Nullable indicates that the value encoded or decoded can be either the
// underlying type or a null value
func (o *SchemaOptions) Nullable() bool {
	return o.nullable
}

// SetNullable sets the nullable flag
func (o *SchemaOptions) SetNullable(n bool) {
	o.nullable = n
}

// WeakDecoding indicates that the schema should be more lenient when decoding
// (i.e. decoding a boolean to a string). See README for details.
func (o *SchemaOptions) WeakDecoding() bool {
	return o.weakDecoding
}

// SetWeakDecoding sets the weak decoding flag
func (o *SchemaOptions) SetWeakDecoding(w bool) {
	o.weakDecoding = w
}
