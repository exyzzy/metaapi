package metasql

import (
	"errors"
	"fmt"
	"log"

	pluralize "github.com/gertd/go-pluralize"
	lex "github.com/timtadh/lexmachine"
)

type NextAction struct {
	State int
	Fn    func(*StateMachine, *lex.Token)
}

func getColumn(sm *StateMachine) *Column {
	if len(sm.Tables) > 0 {
		table := &(sm.Tables[len(sm.Tables)-1])
		if len(table.Columns) > 0 {
			return &(table.Columns[len(table.Columns)-1])
		} else {
			return nil
		}
	} else {
		return nil
	}

}

func InitState(fname string) *StateMachine {
	sm := new(StateMachine)
	sm.FName = fname
	sm.CurState = 1
	plural = pluralize.NewClient()
	return sm
}

func error_state(sm *StateMachine, token *lex.Token) {
	//no state found
	log.Panic("Error in SQL Syntax!")
}

func nop(sm *StateMachine, token *lex.Token) {
	//nop
}

func create_table(sm *StateMachine, token *lex.Token) {
	sm.Tables = append(sm.Tables, Table{})
}

func table_name(sm *StateMachine, token *lex.Token) {
	if len(sm.Tables) > 0 {
		sm.Tables[len(sm.Tables)-1].Name = string(token.Lexeme)
		t := lowerize(string(token.Lexeme))
		p := plural.Plural(t)
		if t != p {
			fmt.Println("WARNING: ALL TABLE NAMES SHOULD BE PLURAL: ", t, "==>", p)
		}
	}
}

func column_name(sm *StateMachine, token *lex.Token) {
	if len(sm.Tables) > 0 {
		table := &(sm.Tables[len(sm.Tables)-1])
		table.Columns = append(table.Columns, Column{})
		table.Columns[len(table.Columns)-1].Name = string(token.Lexeme)
		table.Columns[len(table.Columns)-1].Not = false
		table.Columns[len(table.Columns)-1].Null = true //default
	}
}

func data_type(sm *StateMachine, token *lex.Token) {
	column := getColumn(sm)
	column.Type = Tokens[token.Type]
}

func some_ref(sm *StateMachine, token *lex.Token) {
	column := getColumn(sm)
	column.Ref = true
}

func col_not(sm *StateMachine, token *lex.Token) {
	column := getColumn(sm)
	column.Not = true
}

func col_null(sm *StateMachine, token *lex.Token) {
	column := getColumn(sm)
	if column.Not {
		column.Null = false
	} else {
		column.Null = true
	}
}

func col_primary(sm *StateMachine, token *lex.Token) {
	column := getColumn(sm)
	column.Null = false
}

func some_stuff(sm *StateMachine, token *lex.Token) {
	//nop
}
func end_table(sm *StateMachine, token *lex.Token) {
	//nop
}

func appendQuery(sm *StateMachine, st string) {
	if len(sm.Tables) > 0 {
		(&(sm.Tables[len(sm.Tables)-1])).Query += st + " "
	}
}

func printQuery(sm *StateMachine) {
	if len(sm.Tables) > 0 {
		fmt.Println("query: ", (&(sm.Tables[len(sm.Tables)-1])).Query, " <<")
	}
}

func ProcessState(sm *StateMachine, token *lex.Token) (err error) {

	//State Machine, format is:
	//"CurState, InToken": {NextState, FunctionToCall}

	stateMap := map[string]NextAction{
		"Error":         {0, error_state},
		"1,CREATE":      {2, create_table},
		"2,TABLE":       {3, nop},
		"2,ID":          {2, some_stuff},
		"3,IF":          {4, nop},
		"4,NOT":         {5, nop},
		"5,EXISTS":      {3, nop},
		"3,ID":          {6, table_name},
		"6,(":           {7, nop},
		"7,ID":          {8, column_name},
		"7,UUID":        {8, column_name},
		"8,BOOLEAN":     {9, data_type},
		"8,BOOL":        {9, data_type},
		"8,CHARID":      {9, data_type},
		"8,VARCHARID":   {9, data_type},
		"8,TEXT":        {9, data_type},
		"8,SMALLINT":    {9, data_type},
		"8,INT":         {9, data_type},
		"8,INTEGER":     {9, data_type},
		"8,BIGINT":      {9, data_type},
		"8,SMALLSERIAL": {9, data_type},
		"8,SERIAL":      {9, data_type},
		"8,BIGSERIAL":   {9, data_type},
		"8,FLOATID":     {9, data_type},
		"8,REAL":        {9, data_type},
		"8,FLOAT8":      {9, data_type},
		"8,DECIMAL":     {9, data_type},
		"8,NUMERIC":     {9, data_type},
		"8,NUMERICID":   {9, data_type},
		"8,DOUBLE":      {10, nop},
		"10,PRECISION":  {9, data_type},
		"8,DATE":        {9, data_type},
		"8,TIME":        {9, data_type},
		"8,TIMESTAMPTZ": {9, data_type},
		"8,TIMESTAMP":   {9, data_type},
		"8,INTERVAL":    {9, data_type},
		"8,JSON":        {9, data_type},
		"8,JSONB":       {9, data_type},
		"8,UUID":        {9, data_type},
		"9,,":           {7, nop},
		"9,)":           {11, nop},
		"9,REFID":       {9, some_ref},
		"9,NOT":         {9, col_not},
		"9,NULL":        {9, col_null},
		"9,PRIMARY":     {9, col_primary},
		"9,ID":          {9, some_stuff},
		"11,;":          {1, end_table},
		"11,ID":         {11, some_stuff},
	}

	mapStr := fmt.Sprintf("%d,%s", sm.CurState, Tokens[token.Type])
	nextState := stateMap[mapStr]
	//map zeros all fields of struct if not found
	if nextState.State == 0 {
		printQuery(sm)
		err = errors.New("Syntax Error: " + Tokens[token.Type])
		return
	}
	sm.CurState = nextState.State
	nextState.Fn(sm, token)
	appendQuery(sm, string(token.Lexeme))
	return nil
}
