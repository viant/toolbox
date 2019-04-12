package toolbox

import (
	"bufio"
	"io"
	"unicode"
)

//IsASCIIText return true if supplied string does not have binary data
func IsASCIIText(candidate string) bool {
	for _, r := range candidate {
		if r == '\n' || r == '\r' || r == '\t' {
			continue
		}
		if r > unicode.MaxASCII || !unicode.IsPrint(r) || r == '`' {
			return false
		}
	}
	return true
}

//IsPrintText return true if all candidate characters are printable (unicode.IsPrintText)
func IsPrintText(candidate string) bool {
	for _, r := range candidate {
		if !unicode.IsPrint(r) {
			if r == '\n' || r == '\r' || r == '\t' || r == '`' {
				continue
			}
			return false
		}
	}
	return true
}

//TerminatedSplitN split supplied text into n fragmentCount, each terminated with supplied terminator
func TerminatedSplitN(text string, fragmentCount int, terminator string) []string {
	var result = make([]string, 0)
	if fragmentCount == 0 {
		fragmentCount = 1
	}
	fragmentSize := len(text) / fragmentCount
	lowerBound := 0
	for i := fragmentSize - 1; i < len(text); i++ {
		isLast := i+1 == len(text)
		isAtLeastOfFragmentSize := i-lowerBound >= fragmentSize
		isNewLine := string(text[i:i+len(terminator)]) == terminator
		if (isAtLeastOfFragmentSize && isNewLine) || isLast {
			result = append(result, string(text[lowerBound:i+1]))
			lowerBound = i + 1
		}
	}
	return result
}

//SplitTextStream divides reader supplied text by number of specified line
func SplitTextStream(reader io.Reader, writerProvider func() io.WriteCloser, elementCount int) error {
	scanner := bufio.NewScanner(reader)
	var writer io.WriteCloser
	counter := 0
	var err error
	if elementCount == 0 {
		elementCount = 1
	}
	for scanner.Scan() {

		if writer == nil {
			writer = writerProvider()
		}
		data := scanner.Bytes()
		if err = scanner.Err(); err != nil {
			return err
		}

		if counter > 0 {
			if _, err = writer.Write([]byte{'\n'}); err != nil {
				return err
			}
		}
		if _, err = writer.Write(data); err != nil {
			return err
		}
		counter++
		if counter == elementCount {
			if err := writer.Close(); err != nil {
				return err
			}
			counter = 0
			writer = nil
		}
	}
	if writer != nil {
		return writer.Close()
	}
	return nil
}

const (
	CaseUpper = iota
	CaseLower
	CaseUpperCamel
	CaseLowerCamel
	CaseUpperUnderscore
	CaseLowerUnderscore
)

//ToCaseFormat format text,  from, to are const:  CaseLower, CaseUpperCamel,  CaseLowerCamel,  CaseUpperUnderscore,  CaseLowerUnderscore,
func ToCaseFormat(text string, from, to int) string {
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
			if unicode.IsUpper(r) && fromCamel {
				if toUnserscore {
					result = append(result, underscore)
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
