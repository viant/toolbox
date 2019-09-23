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

func Test_NewCustomIdMatcher(t *testing.T) {
	{
		matcher := toolbox.NewCustomIdMatcher("$")
		assert.Equal(t, 5, matcher.Match("Z $Abcf", 2))
		assert.Equal(t, 1, matcher.Match("Z Abcf", 0))
		assert.Equal(t, 0, matcher.Match("### ##", 0))
	}
	matcher := toolbox.NewCustomIdMatcher("_", "(", ")")
	assert.Equal(t, 1, matcher.Match(" v_sc()", 6))

}

func Test_NewSequenceMatcher(t *testing.T) {
	matcher := toolbox.NewSequenceMatcher("&&", "||")
	assert.Equal(t, 2, matcher.Match("123", 1))
	assert.Equal(t, 4, matcher.Match("123 && 123", 0))
}

func Test_NewSingleSequenceMatcher(t *testing.T) {
	matcher := toolbox.NewSequenceMatcher("&")
	assert.Equal(t, 0, matcher.Match("123", 1))
	assert.Equal(t, 5, matcher.Match("12345&3", 0))

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
		matcher := toolbox.BodyMatcher{Begin: "{", End: "}"}
		var text = " {    {  \n}     }  "
		pos := matcher.Match(text, 1)
		assert.Equal(t, 16, pos)
	}
	{
		matcher := toolbox.BodyMatcher{Begin: "begin", End: "end"}
		var text = " begin  {  \n}     end  "
		pos := matcher.Match(text, 1)
		assert.Equal(t, 20, pos)
	}
}

func TestBlockMatcher(t *testing.T) {
	{
		matcher := toolbox.NewBlockMatcher(false, "begin", "end;", []string{"CASE"}, []string{"END IF"})
		text := ` TRIGGER users_before_insert
BEFORE INSERT ON users
FOR EACH ROW
BEGIN
SELECT users_seq.NEXTVAL
INTO   :new.id
FROM   dual;
END;

INSERT INTO DUMMY(ID, NAME) VALUES(2, 'xyz');

`
		matcher.Match(text, 65)
	}
	{
		matcher := toolbox.BlockMatcher{
			CaseSensitive:      false,
			SequenceStart:      "begin",
			SequenceTerminator: "end",
			NestedSequences:    []string{"case"},
			IgnoredTerminators: []string{"end if"},
		}
		text := "\n\n" +
			"BEgin\n" +
			"IF get_version()=20\n" +
			"select *\n" +
			"from table\n" +
			"where color = case inventory when 1 then 'brown' when 2 then 'red' END;\n" +
			"END IF\n" +
			"END;"
		pos := matcher.Match(text, 2)
		assert.Equal(t, 128, pos)
	}
	{
		matcher := toolbox.BlockMatcher{
			CaseSensitive:      false,
			SequenceStart:      "begin",
			SequenceTerminator: "end;",
			NestedSequences:    []string{"case"},
		}
		text := "bEgIn case 123deabc then 22 End; End;"
		pos := matcher.Match(text, 0)
		assert.Equal(t, 37, pos)

		matcher.CaseSensitive = true
		pos = matcher.Match(text, 0)
		assert.Equal(t, 0, pos)

		matcher.SequenceTerminator = "End;"
		matcher.SequenceStart = "bEgIn"
		pos = matcher.Match(text, 0)
		assert.Equal(t, 37, pos)
	}

}
