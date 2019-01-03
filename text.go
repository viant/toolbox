package toolbox

import "unicode"

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
		isAtLeastOfFragementSize := i-lowerBound >= fragmentSize
		isNewLine := string(text[i:i+len(terminator)]) == terminator
		if (isAtLeastOfFragementSize && isNewLine) || isLast {
			result = append(result, string(text[lowerBound:i+1]))
			lowerBound = i + 1
		}
	}
	return result
}
