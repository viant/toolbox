package toolbox_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/viant/toolbox"
)

func TestNewTokenizer(t *testing.T) {
	tokenizer := toolbox.NewTokenizer("Z Abcf",
		0,
		-1,
		map[int]toolbox.Matcher{
			101: toolbox.KeywordMatcher{"Abc", true},
			201: toolbox.CharactersMatcher{Chars: " \n\t"},
			102: toolbox.LiteralMatcher{},
		},
	)

	assert.Equal(t, 102, tokenizer.Nexts(101, 201, 102).Token)
	assert.Equal(t, 201, tokenizer.Nexts(101, 201, 102).Token)
	assert.Equal(t, 101, tokenizer.Nexts(101, 201, 102).Token)
	assert.Equal(t, 102, tokenizer.Nexts(101, 201, 102).Token)

}

func TestMatchKeyword(t *testing.T) {
	matcher := toolbox.KeywordMatcher{"Abc", true}
	assert.Equal(t, 3, matcher.Match("Z Abcf", 2))
	assert.Equal(t, 0, matcher.Match("Z Abcf", 0))
	assert.Equal(t, 3, matcher.Match("Z Abc", 2))

}

func TestMatchWhitespace(t *testing.T) {
	matcher := toolbox.CharactersMatcher{Chars: " \n\t"}
	assert.Equal(t, 0, matcher.Match("1, 2, 3", 0))
	assert.Equal(t, 2, matcher.Match("1, \t2, 3", 2))

}

func TestLiteralMatcher(t *testing.T) {
	matcher := toolbox.LiteralMatcher{}
	assert.Equal(t, 0, matcher.Match(" abc ", 0))
	assert.Equal(t, 4, matcher.Match(" a1bc", 1))

}

func TestEOFMatcher(t *testing.T) {
	matcher := toolbox.EOFMatcher{}
	assert.Equal(t, 0, matcher.Match(" abc ", 0))
	assert.Equal(t, 1, matcher.Match(" a1bc", 4))
}

func TestKeywordsMatcher(t *testing.T) {
	{
		matcher := toolbox.KeywordsMatcher{Keywords: []string{"ab", "xy"},
			CaseSensitive: false}
		assert.Equal(t, 2, matcher.Match(" abcde", 1))
		assert.Equal(t, 0, matcher.Match(" abcde", 0))
	}
	{
		matcher := toolbox.KeywordsMatcher{Keywords: []string{"AB", "xy"},
			CaseSensitive: true}
		assert.Equal(t, 2, matcher.Match(" ABcde", 1))
		assert.Equal(t, 0, matcher.Match("abcde", 0))
	}
}


func TestBodyMatcher(t *testing.T) {
	{
		matcher := toolbox.BodyMatcher{Begin:"{", End:"}"}
		var text = " {    {  \n}     }  "
		assert.Equal(t, 15, matcher.Match(text, 1))
	}
}