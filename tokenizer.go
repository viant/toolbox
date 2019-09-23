package toolbox

import (
	"fmt"
	"strings"
	"unicode"
)

//Matcher represents a matcher, that matches input from offset position, it returns number of characters matched.
type Matcher interface {
	//Match matches input starting from offset, it return number of characters matched
	Match(input string, offset int) (matched int)
}

//Token a matchable input
type Token struct {
	Token   int
	Matched string
}

//Tokenizer represents a token scanner.
type Tokenizer struct {
	matchers       map[int]Matcher
	Input          string
	Index          int
	InvalidToken   int
	EndOfFileToken int
}

//Nexts matches the first of the candidates
func (t *Tokenizer) Nexts(candidates ...int) *Token {
	for _, candidate := range candidates {
		result := t.Next(candidate)
		if result.Token != t.InvalidToken {
			return result

		}
	}
	return &Token{t.InvalidToken, ""}
}

//Next tries to match a candidate, it returns token if imatching is successful.
func (t *Tokenizer) Next(candidate int) *Token {
	offset := t.Index
	if !(offset < len(t.Input)) {
		return &Token{t.EndOfFileToken, ""}
	}

	if candidate == t.EndOfFileToken {
		return &Token{t.InvalidToken, ""}
	}
	if matcher, ok := t.matchers[candidate]; ok {
		matchedSize := matcher.Match(t.Input, offset)
		if matchedSize > 0 {
			t.Index = t.Index + matchedSize
			return &Token{candidate, t.Input[offset : offset+matchedSize]}
		}

	} else {
		panic(fmt.Sprintf("failed to lookup matcher for %v", candidate))
	}
	return &Token{t.InvalidToken, ""}
}

//NewTokenizer creates a new NewTokenizer, it takes input, invalidToken, endOfFileToeken, and matchers.
func NewTokenizer(input string, invalidToken int, endOfFileToken int, matcher map[int]Matcher) *Tokenizer {
	return &Tokenizer{
		matchers:       matcher,
		Input:          input,
		Index:          0,
		InvalidToken:   invalidToken,
		EndOfFileToken: endOfFileToken,
	}
}

//CharactersMatcher represents a matcher, that matches any of Chars.
type CharactersMatcher struct {
	Chars string //characters to be matched
}

//Match matches any characters defined in Chars in the input, returns 1 if character has been matched
func (m CharactersMatcher) Match(input string, offset int) int {
	var matched = 0
	if offset >= len(input) {
		return matched
	}
outer:
	for _, r := range input[offset:] {
		for _, candidate := range m.Chars {
			if candidate == r {
				matched++
				continue outer
			}
		}
		break
	}
	return matched
}

//NewCharactersMatcher creates a new character matcher
func NewCharactersMatcher(chars string) Matcher {
	return &CharactersMatcher{Chars: chars}
}

//EOFMatcher represents end of input matcher
type EOFMatcher struct {
}

//Match returns 1 if end of input has been reached otherwise 0
func (m EOFMatcher) Match(input string, offset int) int {
	if offset+1 == len(input) {
		return 1
	}
	return 0
}

//IntMatcher represents a matcher that finds any int in the input
type IntMatcher struct{}

//Match matches a literal in the input, it returns number of character matched.
func (m IntMatcher) Match(input string, offset int) int {
	var matched = 0
	if offset >= len(input) {
		return matched
	}
	for _, r := range input[offset:] {
		if !unicode.IsDigit(r) {
			break
		}
		matched++
	}
	return matched
}

//NewIntMatcher returns a new integer matcher
func NewIntMatcher() Matcher {
	return &IntMatcher{}
}

var dotRune = rune('.')
var underscoreRune = rune('_')

//LiteralMatcher represents a matcher that finds any literals in the input
type LiteralMatcher struct{}

//Match matches a literal in the input, it returns number of character matched.
func (m LiteralMatcher) Match(input string, offset int) int {
	var matched = 0
	if offset >= len(input) {
		return matched
	}
	for i, r := range input[offset:] {
		if i == 0 {
			if !unicode.IsLetter(r) {
				break
			}
		} else if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == dotRune || r == underscoreRune) {
			break
		}
		matched++
	}
	return matched
}

//LiteralMatcher represents a matcher that finds any literals in the input
type IdMatcher struct{}

//Match matches a literal in the input, it returns number of character matched.
func (m IdMatcher) Match(input string, offset int) int {
	var matched = 0
	if offset >= len(input) {
		return matched
	}
	for i, r := range input[offset:] {
		if i == 0 {
			if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
				break
			}
		} else if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == dotRune || r == underscoreRune) {
			break
		}
		matched++
	}
	return matched
}

//SequenceMatcher represents a matcher that finds any sequence until find provided terminators
type SequenceMatcher struct {
	Terminators            []string
	CaseSensitive          bool
	matchAllIfNoTerminator bool
	runeTerminators        []rune
}

func (m *SequenceMatcher) hasTerminator(candidate string) bool {
	var candidateLength = len(candidate)
	for _, terminator := range m.Terminators {
		terminatorLength := len(terminator)
		if len(terminator) > candidateLength {
			continue
		}
		if !m.CaseSensitive {
			if strings.ToLower(terminator) == strings.ToLower(string(candidate[:terminatorLength])) {
				return true
			}
		}
		if terminator == string(candidate[:terminatorLength]) {
			return true
		}
	}
	return false
}

//Match matches a literal in the input, it returns number of character matched.
func (m *SequenceMatcher) Match(input string, offset int) int {
	var matched = 0
	hasTerminator := false
	if offset >= len(input) {
		return matched
	}
	if len(m.runeTerminators) > 0 {
		return m.matchSingleTerminator(input, offset)
	}
	var i = 0
	for ; i < len(input)-offset; i++ {
		if m.hasTerminator(string(input[offset+i:])) {
			hasTerminator = true
			break
		}
	}
	if !hasTerminator && !m.matchAllIfNoTerminator {
		return 0
	}
	return i
}

func (m *SequenceMatcher) matchSingleTerminator(input string, offset int) int {
	matched := 0
	hasTerminator := false
outer:
	for i, r := range input[offset:] {
		for _, terminator := range m.runeTerminators {
			terminator = unicode.ToLower(terminator)
			if m.CaseSensitive {
				r = unicode.ToLower(r)
				terminator = unicode.ToLower(terminator)
			}
			if r == terminator {
				hasTerminator = true
				matched = i
				break outer
			}
		}

	}
	if !hasTerminator && !m.matchAllIfNoTerminator {
		return 0
	}
	return matched
}

//NewSequenceMatcher creates a new matcher that finds all sequence until find at least one of the provided terminators
func NewSequenceMatcher(terminators ...string) Matcher {
	result := &SequenceMatcher{
		matchAllIfNoTerminator: true,
		Terminators:            terminators,
		runeTerminators:        []rune{},
	}
	for _, terminator := range terminators {
		if len(terminator) != 1 {
			result.runeTerminators = []rune{}
			break
		}
		result.runeTerminators = append(result.runeTerminators, rune(terminator[0]))
	}
	return result
}

//NewTerminatorMatcher creates a new matcher that finds any sequence until find at least one of the provided terminators
func NewTerminatorMatcher(terminators ...string) Matcher {
	result := &SequenceMatcher{
		Terminators:     terminators,
		runeTerminators: []rune{},
	}
	for _, terminator := range terminators {
		if len(terminator) != 1 {
			result.runeTerminators = []rune{}
			break
		}
		result.runeTerminators = append(result.runeTerminators, rune(terminator[0]))
	}
	return result
}

//remainingSequenceMatcher represents a matcher that matches all reamining input
type remainingSequenceMatcher struct{}

//Match matches a literal in the input, it returns number of character matched.
func (m *remainingSequenceMatcher) Match(input string, offset int) (matched int) {
	return len(input) - offset
}

//Creates a matcher that matches all remaining input
func NewRemainingSequenceMatcher() Matcher {
	return &remainingSequenceMatcher{}
}

//CustomIdMatcher represents a matcher that finds any literals with additional custom set of characters in the input
type customIdMatcher struct {
	Allowed map[rune]bool
}

func (m *customIdMatcher) isValid(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return true
	}
	return m.Allowed[r]
}

//Match matches a literal in the input, it returns number of character matched.
func (m *customIdMatcher) Match(input string, offset int) int {
	var matched = 0
	if offset >= len(input) {
		return matched
	}
	for _, r := range input[offset:] {
		if !m.isValid(r) {
			break
		}
		matched++
	}
	return matched
}

//NewCustomIdMatcher creates new custom matcher
func NewCustomIdMatcher(allowedChars ...string) Matcher {
	var result = &customIdMatcher{
		Allowed: make(map[rune]bool),
	}
	if len(allowedChars) == 1 && len(allowedChars[0]) > 0 {
		for _, allowed := range allowedChars[0] {
			result.Allowed[rune(allowed)] = true
		}
	}
	for _, allowed := range allowedChars {
		result.Allowed[rune(allowed[0])] = true
	}
	return result
}

//LiteralMatcher represents a matcher that finds any literals in the input
type BodyMatcher struct {
	Begin string
	End   string
}

//Match matches a literal in the input, it returns number of character matched.
func (m *BodyMatcher) Match(input string, offset int) (matched int) {
	beginLen := len(m.Begin)
	endLen := len(m.End)
	uniEnclosed := m.Begin == m.End

	if offset+beginLen >= len(input) {
		return 0
	}
	if input[offset:offset+beginLen] != m.Begin {
		return 0
	}
	var depth = 1
	var i = 1
	for ; i < len(input)-offset; i++ {
		canCheckEnd := offset+i+endLen <= len(input)
		if !canCheckEnd {
			return 0
		}
		if !uniEnclosed {
			canCheckBegin := offset+i+beginLen <= len(input)
			if canCheckBegin {
				if string(input[offset+i:offset+i+beginLen]) == m.Begin {
					depth++
				}
			}
		}
		if string(input[offset+i:offset+i+endLen]) == m.End {
			depth--
		}
		if depth == 0 {
			i += endLen
			break
		}
	}
	return i
}

//NewBodyMatcher creates a new body matcher
func NewBodyMatcher(begin, end string) Matcher {
	return &BodyMatcher{Begin: begin, End: end}
}

// Parses SQL Begin End blocks
func NewBlockMatcher(caseSensitive bool, sequenceStart string, sequenceTerminator string, nestedSequences []string, ignoredTerminators []string) Matcher {
	return &BlockMatcher{
		CaseSensitive:      caseSensitive,
		SequenceStart:      sequenceStart,
		SequenceTerminator: sequenceTerminator,
		NestedSequences:    nestedSequences,
		IgnoredTerminators: ignoredTerminators,
	}
}

type BlockMatcher struct {
	CaseSensitive      bool
	SequenceStart      string
	SequenceTerminator string
	NestedSequences    []string
	IgnoredTerminators []string
}

func (m *BlockMatcher) Match(input string, offset int) (matched int) {

	sequenceStart := m.SequenceStart
	terminator := m.SequenceTerminator
	nestedSequences := m.NestedSequences
	ignoredTerminators := m.IgnoredTerminators
	in := input

	starterLen := len(sequenceStart)
	terminatorLen := len(terminator)

	if !m.CaseSensitive {
		sequenceStart = strings.ToLower(sequenceStart)
		terminator = strings.ToLower(terminator)
		for i, seq := range nestedSequences {
			nestedSequences[i] = strings.ToLower(seq)
		}
		for i, term := range ignoredTerminators {
			ignoredTerminators[i] = strings.ToLower(term)
		}
		in = strings.ToLower(input)
	}

	if offset+starterLen >= len(in) {
		return 0
	}
	if in[offset:offset+starterLen] != sequenceStart {
		return 0
	}
	var depth = 1
	var i = 1
	for ; i < len(in)-offset; i++ {
		canCheckEnd := offset+i+terminatorLen <= len(in)
		if !canCheckEnd {
			return 0
		}
		canCheckBegin := offset+i+starterLen <= len(in)
		if canCheckBegin {
			beginning := in[offset+i : offset+i+starterLen]

			if beginning == sequenceStart {
				depth++
			} else {
				for _, nestedSeq := range nestedSequences {
					nestedLen := len(nestedSeq)
					if offset+i+nestedLen >= len(in) {
						continue
					}

					beginning := in[offset+i : offset+i+nestedLen]
					if beginning == nestedSeq {
						depth++
						break
					}
				}
			}
		}
		ignored := false
		for _, ignoredTerm := range ignoredTerminators {
			termLen := len(ignoredTerm)
			if offset+i+termLen >= len(in) {
				continue
			}

			ending := in[offset+i : offset+i+termLen]
			if ending == ignoredTerm {
				ignored = true
				break
			}
		}
		if !ignored && in[offset+i:offset+i+terminatorLen] == terminator && unicode.IsSpace(rune(in[offset+i-1])) {
			depth--
		}
		if depth == 0 {
			i += terminatorLen
			break
		}
	}
	return i
}

//KeywordMatcher represents a keyword matcher
type KeywordMatcher struct {
	Keyword       string
	CaseSensitive bool
}

//Match matches keyword in the input,  it returns number of character matched.
func (m KeywordMatcher) Match(input string, offset int) (matched int) {
	if !(offset+len(m.Keyword)-1 < len(input)) {
		return 0
	}
	if m.CaseSensitive {
		if input[offset:offset+len(m.Keyword)] == m.Keyword {
			return len(m.Keyword)
		}
	} else {
		if strings.ToLower(input[offset:offset+len(m.Keyword)]) == strings.ToLower(m.Keyword) {
			return len(m.Keyword)
		}
	}
	return 0
}

//KeywordsMatcher represents a matcher that finds any of specified keywords in the input
type KeywordsMatcher struct {
	Keywords      []string
	CaseSensitive bool
}

//Match matches any specified keyword,  it returns number of character matched.
func (m KeywordsMatcher) Match(input string, offset int) (matched int) {
	for _, keyword := range m.Keywords {
		if len(input)-offset < len(keyword) {
			continue
		}
		if m.CaseSensitive {
			if input[offset:offset+len(keyword)] == keyword {
				return len(keyword)
			}
		} else {
			if strings.ToLower(input[offset:offset+len(keyword)]) == strings.ToLower(keyword) {
				return len(keyword)
			}
		}
	}
	return 0
}

//NewKeywordsMatcher returns a matcher for supplied keywords
func NewKeywordsMatcher(caseSensitive bool, keywords ...string) Matcher {
	return &KeywordsMatcher{CaseSensitive: caseSensitive, Keywords: keywords}
}

//IllegalTokenError represents illegal token error
type IllegalTokenError struct {
	Illegal  *Token
	Message  string
	Expected []int
	Position int
}

func (e *IllegalTokenError) Error() string {
	return fmt.Sprintf("%v; illegal token at %v [%v], expected %v, but had: %v", e.Message, e.Position, e.Illegal.Matched, e.Expected, e.Illegal.Token)
}

//NewIllegalTokenError create a new illegal token error
func NewIllegalTokenError(message string, expected []int, position int, found *Token) error {
	return &IllegalTokenError{
		Message:  message,
		Illegal:  found,
		Expected: expected,
		Position: position,
	}
}

//ExpectTokenOptionallyFollowedBy returns second matched token or error if first and second group was not matched
func ExpectTokenOptionallyFollowedBy(tokenizer *Tokenizer, first int, errorMessage string, second ...int) (*Token, error) {
	_, _ = ExpectToken(tokenizer, "", first)
	return ExpectToken(tokenizer, errorMessage, second...)
}

//ExpectToken returns the matched token or error
func ExpectToken(tokenizer *Tokenizer, errorMessage string, candidates ...int) (*Token, error) {
	token := tokenizer.Nexts(candidates...)
	hasMatch := HasSliceAnyElements(candidates, token.Token)
	if !hasMatch {
		return nil, NewIllegalTokenError(errorMessage, candidates, tokenizer.Index, token)
	}
	return token, nil
}
