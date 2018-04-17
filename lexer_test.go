package mohajer

import (
	"fmt"
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
	}
)

func TestBasicLexer(t *testing.T) {
	for i := range allLexeme {
		l := lex(allLexeme[i].input)
		require.NotNil(t, l)
		cnt := 0
		fmt.Println("---")
		for j := range l.items {
			fmt.Println(j)
			assert.Equal(t, allLexeme[i].lexeme[cnt], j.typ)
			cnt++
		}
	}
}
