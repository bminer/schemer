package schemer

import (
	"regexp"
	"strings"
)

// StructTagName is the Go struct tag name used by schemer
const StructTagName string = "schemer"

// StructTag represents information that can be parsed from a schemer struct tag
type StructTag struct {
	FieldAliases    []string // list of names of fields. This comes from the name of the field in the struct, and also struct tags
	FieldAliasesSet bool     // if true, struct tags are present on this field
	Nullable        bool     // is this field nullable?
	NullableSet     bool     // if true, then the null struct tags is present on this field
	WeakDecoding    bool     // does this field allow weak decoding?
	WeakDecodingSet bool     // if true, weak decoding was present in the struct tag for this field
}

const tagAlias = `[A-Za-z0-9_]+`
const tagAliases = `\[(` + tagAlias + `(?:,` + tagAlias + `)*)\]`
const tagOpt = `(?:weak|null)`

// ^(\-|([A-Za-z0-9_]+)|\[([A-Za-z0-9_]+(?:,[A-Za-z0-9_]+)*)\])?(,\!?(?:weak|null))*$
// Note: non-capturing groups use regex syntax (?:   ...   )
// Group 1 = Alias / Aliases
// Group 2 = Single alias
// Group 3 = Multiple aliases (comma-delimited)
// Group 4 = Options (comma-delimited; prefixed with a comma)
var structTagRegex *regexp.Regexp = regexp.MustCompile(
	`^(\-|(` + tagAlias + `)|` + tagAliases + `)?(,\!?` + tagOpt + `)*$`,
)

// ParseStructTag parses a struct tag string and returns a decoded StructTag
// the format of the tag must be:
// tag := alias? ("," option)*
// alias := "-" |
//          identifier |
//          "[" identifier ("," identifier)* "]"
// option := "!" ? ( "weak" | "null" )
// If the
func ParseStructTag(s string) (tag StructTag) {
	// See `structTagRegex` documentation
	match := structTagRegex.FindStringSubmatch(s)
	if match == nil || len(match) != 5 {
		return
	}
	singleAlias, aliasesStr, optsStr := match[2], match[3], match[4]

	// Check to see if we are skipping this field
	if match[1] == "-" {
		tag.FieldAliasesSet = true
		return
	}

	// Update aliases
	if singleAlias != "" {
		tag.FieldAliases = []string{singleAlias}
		tag.FieldAliasesSet = true
	} else if aliasesStr != "" {
		tag.FieldAliases = strings.Split(aliasesStr, ",")
		tag.FieldAliasesSet = true
	}

	if optsStr != "" {
		// Parse `optsStr` by removing leading comma and splitting
		opts := strings.Split(optsStr[1:], ",")

		// Update options
		for _, opt := range opts {
			switch opt {
			case "null":
				tag.Nullable = true
				tag.NullableSet = true
			case "!null":
				tag.Nullable = false
				tag.NullableSet = true
			case "weak":
				tag.WeakDecoding = true
				tag.WeakDecodingSet = true
			case "!weak":
				tag.WeakDecoding = false
				tag.WeakDecodingSet = true
			}
		}
	}
	return
}
