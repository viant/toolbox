/*
 *
 *
 * Copyright 2012-2016 Viant.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 *  use this file except in compliance with the License. You may obtain a copy of
 *  the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 *  License for the specific language governing permissions and limitations under
 *  the License.
 *
 */
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
