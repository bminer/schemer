package schemer

import (
	"fmt"
	"strings"
)

// SchemerTagOptions represents information that can be read from struct field tags
type SchemerTagOptions struct {
	FieldAliases []string

	Nullable     bool
	WeakDecoding bool

	ShouldSkip bool
}

// ParseStructTag tags a tagname as a string, parses it, and populates SchemerTagOptions
// the format of the tag must be:
// tag := (alias)?("," option)*
// alias := identifier
//			"["identifier(","identifier)*"]"
//	option := "weak", "null", "not null"
func (s *SchemerTagOptions) ParseStructTag(tagStr string) error {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error parsing struct tag:", err)
		}
	}()

	tagStr = strings.Trim(tagStr, " ")
	if len([]rune(tagStr)) == 0 {
		return nil
	}

	// special case meaning to skip this field
	if tagStr == "-" {
		s.ShouldSkip = true
		return nil
	}

	// if first part has a "]", then extract everything up to there
	// otherwise, extract everything up to the first comma

	var i int
	var aliasStr string
	var optionStr string

	// if the alias portion of the string contains [], then we want to grab everything up
	// to the ] and call that our aliasStr
	if strings.Contains(tagStr, "]") {

		i = strings.Index(tagStr, "]")
		aliasStr = tagStr[0 : i+1]
		tagStr = tagStr[i+1:] // eat off what we just processed
		tagStr = strings.Trim(tagStr, " ")

		if len([]rune(tagStr)) > 0 {

			if !strings.Contains(tagStr, ",") {
				return fmt.Errorf("missing comma after field alias")
			} else {
				// our options are just whatever is left after the comma
				optionStr = tagStr[strings.Index(tagStr, ",")+1:]
				optionStr = strings.Trim(optionStr, " ")
			}

		} else {
			optionStr = ""
		}
	} else {
		i = strings.Index(tagStr, ",")

		if i > 0 {

			// alias string is everything up to the comma
			aliasStr = tagStr[0:i]
			aliasStr = strings.Trim(aliasStr, " ")
			tagStr = tagStr[i+1:] // eat off what we just processed
			tagStr = strings.Trim(tagStr, " ")

			if len([]rune(tagStr)) > 0 {
				// options are everything after the comma
				optionStr = tagStr[strings.Index(tagStr, ",")+1:]
				optionStr = strings.Trim(optionStr, " ")
			} else {
				optionStr = ""
			}
		} else {
			aliasStr = strings.Trim(tagStr, " ")
			optionStr = ""
		}

	}

	// parse aliasStr, and put each field into .FieldAliases
	x := strings.Replace(aliasStr, "[", "", -1)
	y := strings.Replace(x, "]", "", -1)
	s.FieldAliases = strings.Split(y, ",")
	for i, f := range s.FieldAliases {
		s.FieldAliases[i] = strings.Trim(f, " ")
	}

	// parse options, and string and put each option into correct field, such as .Nullable
	s.Nullable = strings.Contains(strings.ToUpper(optionStr), "NUL") &&
		!strings.Contains(strings.ToUpper(optionStr), "!NUL")

	s.WeakDecoding = strings.Contains(strings.ToUpper(optionStr), "WEAK")

	return nil

}
