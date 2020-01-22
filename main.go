package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/exyzzy/metaapi/metasql"

	lex "github.com/timtadh/lexmachine"
)

// Turn on debug prints
var DEBUG = false

func main() {
	sqlPtr := flag.String("sql", "", ".sql input file to parse")
	txtPtr := flag.String("txt", "crud.txt", "go template as .txt file")
	flag.Parse()
	sqlFile := strings.ToLower(*sqlPtr)
	txtFile := strings.ToLower(*txtPtr)

	if (sqlFile == "") || (!strings.HasSuffix(sqlFile, ".sql")) {
		log.Fatal("No .sql File")
	}
	if (txtFile == "") || (!strings.HasSuffix(txtFile, ".txt")) {
		log.Fatal("No .txt File")
	}

	dat, err := ioutil.ReadFile("./" + sqlFile)
	if err != nil {
		log.Fatal(err)
	}

	s, err := metasql.Lexer.Scanner([]byte(dat))
	if err != nil {
		log.Fatal(err)
	}

	sm := metasql.InitState(sqlFile)
	for tok, err, eof := s.Next(); !eof; tok, err, eof = s.Next() {
		if err != nil {
			log.Fatal(err)
		}
		token := tok.(*lex.Token)
		if DEBUG {
			fmt.Printf("%-10v | %-12v | %v:%v-%v:%v\n",
				metasql.Tokens[token.Type],
				string(token.Lexeme),
				token.StartLine,
				token.StartColumn,
				token.EndLine,
				token.EndColumn)
		}
		err = metasql.ProcessState(sm, token)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = metasql.Generate(sm, txtFile)
	if err != nil {
		log.Fatal(err)
	}
	if DEBUG {
		fmt.Printf("Table Capture:\n%+v\n", sm)
	}
}
