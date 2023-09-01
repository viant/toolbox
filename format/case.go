package format

import (
	"fmt"
	"strings"
	"unicode"
)

type Case int

const (
	//CaseUpper represents case format
	CaseUpper = Case(iota)
	CaseLower
	CaseUpperCamel
	CaseLowerCamel
	CaseUpperUnderscore
	CaseLowerUnderscore
)

// NewCase create a new case for supplied name or error, supported in case insensitive form "upperCase", "upper", "u"
func NewCase(name string) (Case, error) {
	switch strings.ToLower(name) {
	case "upper", "u":
		return CaseUpper, nil
	case "lower", "l":
		return CaseLower, nil
	case "lowercamel", "lc":
		return CaseLowerCamel, nil
	case "uppercamel", "uc":
		return CaseUpperCamel, nil
	case "lowerunderscore", "lu":
		return CaseLowerUnderscore, nil
	case "upperunderscore", "uu":
		return CaseUpperUnderscore, nil
	}
	return -1, fmt.Errorf("unsupported case format: %s", name)
}

// String return case format name
func (from Case) String() string {
	switch from {
	case CaseUpper:
		return "Upper"
	case CaseLower:
		return "Lower"
	case CaseUpperCamel:
		return "UpperCamel"
	case CaseLowerCamel:
		return "LowerCamel"
	case CaseUpperUnderscore:
		return "UpperUnderscore"
	case CaseLowerUnderscore:
		return "LowerUnderscore"
	}
	return "UnsupportedCase"
}

// Format converts supplied text from Case to
func (from Case) Format(text string, to Case) string {
	toUpper := false
	toLower := false
	toCamel := false
	toUnserscore := false
	fromCamel := false
	fromUnserscore := false

	switch to {
	case CaseUpper, CaseUpperUnderscore:
		toUpper = true
	case CaseLower, CaseLowerUnderscore:
		toLower = true
	case CaseUpperCamel, CaseLowerCamel:

		toCamel = true

	}
	switch to {
	case CaseUpperUnderscore, CaseLowerUnderscore:
		toUnserscore = true
	}
	switch from {
	case CaseUpperCamel, CaseLowerCamel:
		fromCamel = true
	case CaseUpperUnderscore, CaseLowerUnderscore:
		fromUnserscore = true
	}
	underscore := rune('_')
	var result = make([]rune, 0)
	makeLower := false
	makeUpper := false
	hasUnderscore := false
	for i, r := range text {
		first := i == 0
		if toUpper {
			makeUpper = true
		} else if toLower {
			makeLower = true
		}
		if first {
			if to == CaseLowerCamel {
				r = unicode.ToLower(r)
			} else if to == CaseUpperCamel {
				r = unicode.ToUpper(r)
			}
		} else {
			if fromUnserscore {
				if toCamel {
					if r == underscore {
						hasUnderscore = true
						continue
					}
					if hasUnderscore {
						makeUpper = true
						hasUnderscore = false
					} else {
						makeLower = true
					}
				}
			}
			if unicode.IsUpper(r) {

				if fromCamel {
					if toUnserscore {
						if !(i > 1 && result[len(result)-2] == '_') {
							result = append(result, underscore)
						}
					}
				}
			}
		}

		if makeLower {
			r = unicode.ToLower(r)
		} else if makeUpper {
			r = unicode.ToUpper(r)
		}
		result = append(result, r)
		makeUpper = false
		makeLower = false
	}

	return string(result)
}
