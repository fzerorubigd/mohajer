package mohajer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type lexerTester struct {
	input  string
	lexeme []itemType
}

var (
	allLexeme = []lexerTester{
		{
			`name test

    
create table test

end
`,
			[]itemType{itemName, itemWhiteSpace, itemAlpha, itemNewLine, itemCreate, itemWhiteSpace, itemAlpha, itemWhiteSpace, itemAlpha, itemNewLine, itemEnd, itemNewLine},
		},
		{
			`add column name type:string default : 19 `,
			[]itemType{itemAdd, itemWhiteSpace, itemAlpha, itemWhiteSpace, itemName, itemWhiteSpace, itemAlpha, itemColon, itemAlpha, itemWhiteSpace, itemAlpha, itemWhiteSpace, itemColon, itemWhiteSpace, itemAlpha, itemWhiteSpace},
		},
		{
			`
+add column x default:"string  haha "`,
			[]itemType{itemNewLine, itemSkipDown, itemAdd, itemWhiteSpace, itemAlpha, itemWhiteSpace, itemAlpha, itemWhiteSpace, itemAlpha, itemColon, itemString},
		},
		{
			"add column x `mysql`",
			[]itemType{itemAdd, itemWhiteSpace, itemAlpha, itemWhiteSpace, itemAlpha, itemWhiteSpace, itemOptionTag},
		},
		{
			`add #comment `,
			[]itemType{itemAdd, itemWhiteSpace, itemComment},
		},
		{
			`add "unterminated string `,
			[]itemType{itemAdd, itemWhiteSpace, itemError},
		},
		{
			"add `un-terminated option",
			[]itemType{itemAdd, itemWhiteSpace, itemError},
		},
		{
			`#comment
  #comment again
       #comment
 `,
			[]itemType{itemComment, itemNewLine, itemComment, itemNewLine, itemComment, itemNewLine},
		},
		{
			"add\t\t\n",
			[]itemType{itemAdd, itemNewLine},
		},
		{
			"add\n@",
			[]itemType{itemAdd, itemNewLine, itemError},
		},
		{
			"\n+@",
			[]itemType{itemNewLine, itemSkipDown, itemError},
		},
		{
			"test:@",
			[]itemType{itemAlpha, itemColon, itemError},
		},
		{
			`add "string \" escaped"`,
			[]itemType{itemAdd, itemWhiteSpace, itemString},
		},
		{
			`add "string \t wrong escaped"`,
			[]itemType{itemAdd, itemWhiteSpace, itemError},
		},
		{
			`add@`,
			[]itemType{itemAdd, itemError},
		},
		{
			`add @`,
			[]itemType{itemAdd, itemWhiteSpace, itemError},
		},
	}
)

func TestBasicLexer(t *testing.T) {
	for i := range allLexeme {
		l := lex(allLexeme[i].input)
		require.NotNil(t, l)
		cnt := 0
		for j := range l.items {
			assert.Equal(t, allLexeme[i].lexeme[cnt], j.typ)
			cnt++
		}
		assert.Len(t, allLexeme[i].lexeme, cnt)
	}
}

func TestOtherLexer(t *testing.T) {
	l := lex("name test")
	l.drain()
	assert.Zero(t, l.next())

	assert.Panics(t, func() { l.discard(';') })
	assert.Zero(t, l.nextItem())
}
