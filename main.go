package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/exyzzy/metaapi/metasql"

	lex "github.com/timtadh/lexmachine"
)

// Turn on debug prints
var DEBUG = false

func main() {

	sqlPtr := flag.String("sql", "", ".sql input file to parse")
	txtPtr := flag.String("txt", "api.txt", "go template as .txt file")
	pipePtr := flag.Bool("pipe", false, "use piped generation")

	flag.Parse()

	if flag.NFlag() == 0 {
		fmt.Println(" valid usage is:")
		fmt.Println("  metaapi -sql=yoursql.sql -txt=[api.txt | api_test.txt]")
		fmt.Println("  pipe metaapi -sql=yoursql.sql -pipe=true :: yourproj -txt=yourtemplate.txt -pipe=true")
		os.Exit(1)
	}

	sqlFile := strings.ToLower(*sqlPtr)
	txtFile := strings.ToLower(*txtPtr)

	if (sqlFile != "") && (!strings.HasSuffix(sqlFile, ".sql")) {
		log.Panic("Invalid .sql File")
	}

	if (txtFile != "") && (!strings.HasSuffix(txtFile, ".txt")) {
		log.Panic("Invalid .txt File")
	}
	var sm *(metasql.StateMachine)

	if sqlFile != "" {
		dat, err := ioutil.ReadFile("./" + sqlFile)
		if err != nil {
			log.Panic(err)
		}
		s, err := metasql.Lexer.Scanner([]byte(dat))
		if err != nil {
			log.Panic(err)
		}
		sm = metasql.InitState(sqlFile)
		for tok, err, eof := s.Next(); !eof; tok, err, eof = s.Next() {
			if err != nil {
				log.Panic(err)
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
				log.Panic(err)
			}
		}
	}
	if !*pipePtr {
		err := metasql.Generate(*sm, txtFile)
		if err != nil {
			log.Panic(err)
		}
	} else {
		//send to stdio
		psm, err := json.Marshal(sm)
		if err != nil {
			log.Panic(err)
			return
		}
		fmt.Println(string(psm))
	}
	if DEBUG {
		fmt.Println("sql  File: ", sqlFile)
		fmt.Println("txt  File: ", txtFile)
		fmt.Printf("Table Capture:\n%+v\n", sm)
	}
}
