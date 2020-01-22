//Derived from: https://hackthology.com/writing-a-lexer-in-go-with-lexmachine.html

package metasql

import (
	"strings"

	lex "github.com/timtadh/lexmachine"
	"github.com/timtadh/lexmachine/machines"
)

var Literals []string       // The tokens representing literal strings
var Keywords []string       // The keyword tokens
var Tokens []string         // All of the tokens (including literals and keywords)
var TokenIds map[string]int // A map from the token names to their int ids
var Lexer *lex.Lexer        // The lexer object. Use this to construct a Scanner

// Called at package initialization. Creates the lexer and populates token lists.
func init() {
	initTokens()
	var err error
	Lexer, err = initLexer()
	if err != nil {
		panic(err)
	}
}

func initTokens() {
	Tokens = []string{
		"VARCHARID",
		"CHARID",
		"FLOATID",
		"NUMERICID",
		"REFID",
		"ID",
	}
	Keywords = []string{
		"CREATE",
		"TABLE",
		"IF",
		"NOT",
		"EXISTS",
		"BOOLEAN",
		"BOOL",
		"TEXT",
		"SMALLINT",
		"INTEGER",
		"BIGINT",
		"INT",
		"SMALLSERIAL",
		"BIGSERIAL",
		"SERIAL",
		"REAL",
		"FLOAT8",
		"DECIMAL",
		"NUMERIC",
		"DOUBLE",
		"PRECISION",
		"DATE",
		"TIMESTAMPTZ",
		"TIMESTAMP",
		"TIME",
		"INTERVAL",
		"JSONB",
		"JSON",
		"UUID",
	}
	Literals = []string{
		"(",
		")",
		",",
		";",
	}
	Tokens = append(Tokens, Keywords...)
	Tokens = append(Tokens, Literals...)
	TokenIds = make(map[string]int)
	for i, tok := range Tokens {
		TokenIds[tok] = i
	}
}

// Creates the lexer object and compiles the NFA.
func initLexer() (*lex.Lexer, error) {
	lexer := lex.NewLexer()

	for _, lit := range Literals {
		r := "\\" + strings.Join(strings.Split(lit, ""), "\\")
		lexer.Add([]byte(r), token(lit))
	}
	for _, name := range Keywords {
		lexer.Add([]byte(strings.ToLower(name)), token(name))
	}
	lexer.Add([]byte(`[vV][aA][rR][cC][hH][aA][rR]\([0-9]+\)`), token("VARCHARID"))
	lexer.Add([]byte(`[cC][hH][aA][rR]\([0-9]+\)`), token("CHARID"))
	lexer.Add([]byte(`[fF][lL][oO][aA][tT]\([0-9]+\)`), token("FLOATID"))
	lexer.Add([]byte(`[nN][uU][mM][eE][rR][iI][cC]\([0-9]+,[0-9]+\)`), token("NUMERICID"))
	lexer.Add([]byte(`([a-z]|[A-Z]|_|#|@)([a-z]|[A-Z]|[0-9]|_|#|@|\$)*\(([a-z]|[A-Z]|_|#|@)([a-z]|[A-Z]|[0-9]|_|#|@|\$)*\)`), token("REFID"))
	lexer.Add([]byte(`([a-z]|[A-Z]|_|#|@)([a-z]|[A-Z]|[0-9]|_|#|@|\$)*`), token("ID"))
	lexer.Add([]byte("( |\t|\n|\r)+"), skip)

	err := lexer.Compile()
	if err != nil {
		return nil, err
	}
	return lexer, nil
}

// a lex.Action function which skips the match.
func skip(*lex.Scanner, *machines.Match) (interface{}, error) {
	return nil, nil
}

// a lex.Action function with constructs a Token of the given token type by
// the token type's name.
func token(name string) lex.Action {
	return func(s *lex.Scanner, m *machines.Match) (interface{}, error) {
		return s.Token(TokenIds[name], string(m.Bytes), m), nil
	}
}
