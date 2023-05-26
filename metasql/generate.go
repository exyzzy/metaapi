package metasql

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/exyzzy/metaapi/data"
	pluralize "github.com/gertd/go-pluralize"
)

var plural *pluralize.Client

type Column struct {
	Name      string
	Type      string
	Ref       bool //if true is foreign key
	Not       bool //true if Not seen in column (can refer to Null or Deferrable)
	Null      bool //column is Not Null or Null (default)
	Generated bool //true if GENERATED in column
}

type Table struct {
	Name    string
	Query   string
	Columns []Column
}

type StateMachine struct {
	FName    string //sql fil
	CurState int
	Tables   []Table
}

func ReadSM() (sm *StateMachine, err error) {
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return
	}
	err = json.Unmarshal(bytes, &sm)
	return
}

type FileName struct {
	Name   string
	Prefix bool
}

var fileMap = map[string]FileName{
	"api_test.txt":     {"generated_api_test.go", true},
	"api.txt":          {"generated_api.go", true},
	"v_api_test.txt":   {"generated_v_api_test.go", true},
	"v_api.txt":        {"generated_v_api.go", true},
	"v_route_test.txt": {"generated_v_route_test.go", true},
	"v_route.txt":      {"generated_v_route.go", true},
	"v_readme.txt":     {"README.MD", true},
	"v_tables.txt":     {"tables.html", false},
	"v_tables.vue.txt": {"tables.vue.js", false},
}

// Generate assumes that the primary ID is in the first column (index 0)
func Generate(dta interface{}, txtFile string) error {

	var sm StateMachine
	var prefix string
	var suffix string
	// plural = pluralize.NewClient() //init in parse

	// fmt.Println("GENERATE: ", reflect.TypeOf(dta).Name())
	if reflect.TypeOf(dta).Name() == "StateMachine" {
		sm = dta.(StateMachine)

		if sm.FName == "" {
			return (errors.New("No file name"))
		}
		prefix = sm.FilePrefix()
	}
	if txtFile != "" {
		dir, file := filepath.Split(txtFile)
		dot := strings.Index(file, ".")
		if dot > 0 {
			suffix = file[:dot]
		} else {
			suffix = file
		}

		var dest string
		if fileMap[file].Name != "" {
			if fileMap[file].Prefix {
				dest = prefix + "_"
			}
			dest += fileMap[file].Name
		} else {
			dest = prefix + "_generated_" + suffix + ".go"
		}

		dest = filepath.Join(dir, dest)

		//for -pipe option, instead of data.Asset, use:
		// dat, err := ioutil.ReadFile("./" + txtFile)
		// if err != nil {
		// 	return err
		// }

		dat, err := data.Asset(file)
		if err != nil {
			return err
		}
		return generateFile(dat, &sm, dest)
	}
	return nil
}

func jsonify(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), "_", "")
}

func nullify(c Column) string {
	s := jsonify(c.Name)
	if c.Null {
		s += typeMap[c.Type][2]
	}
	return s
}

func generateFile(templatesrc []byte, data interface{}, dest string) error {
	fmap := template.FuncMap{
		"jsonify": jsonify,
		"nullify": nullify,
	}
	tt := template.Must(template.New("file").Delims("<<", ">>").Funcs(fmap).Parse(string(templatesrc)))
	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	err = tt.Execute(file, data)
	file.Close()
	return err
}

//======== string helpers

func singularize(s string) string {
	if strings.HasSuffix(strings.ToLower(s), "s") {
		return strings.TrimSuffix(strings.ToLower(s), "s")
	} else {
		return strings.ToLower(s)
	}
}

func capitalize(s string) string {
	return strings.Title(s)
}

func lowerize(s string) string {
	return strings.ToLower(s)
}

// turns submitted_at into SubmittedAt, and otherwise capitalizes
func camelize(s string) string {
	return strings.ReplaceAll(strings.Title(strings.ReplaceAll(strings.ToLower(s), "_", " ")), " ", "")
}

func comma(i int, length int) string {
	if i < (length - 1) {
		return ","
	} else {
		return ""
	}
}

//======== template methods

func (sm *StateMachine) ReverseTables() (result []Table) {
	for i := len(sm.Tables) - 1; i >= 0; i-- {
		result = append(result, sm.Tables[i])
	}
	return
}

func (sm *StateMachine) FilePrefix() string {
	dot := strings.Index(sm.FName, ".")
	var prefix string
	if dot > 0 {
		prefix = sm.FName[:dot]
	} else {
		prefix = sm.FName
	}
	return prefix
}

func (sm *StateMachine) Package() string {
	return os.Getenv("GOPACKAGE")
}

func (sm *StateMachine) Import() string {

	var s string

	s += "import (\n\t\"database/sql\"\n"
	s += "\t_ \"github.com/lib/pq\"\n)"
	return s

	// var s string
	// var includeTime bool

	// includeTime = false
	// for _, table := range sm.Tables {
	// 	for _, column := range table.Columns {
	// 		switch column.Type {
	// 		case "DATE", "TIME", "TIMESTAMP", "TIMESTAMPTZ", "INTERVAL":
	// 			if !column.Null {
	// 				includeTime = true
	// 			}
	// 		default:
	// 		}
	// 	}
	// }
	// s += "import (\n\t\"database/sql\"\n"
	// s += "\t_ \"github.com/lib/pq\"\n"

	// if includeTime {
	// 	s += "\t\"time\"\n"
	// }
	// s += ")"
	// return s

}

func (sm *StateMachine) ImportSQL() string {

	var s string
	var includeSQL bool

	includeSQL = false
	for _, table := range sm.Tables {
		for i, column := range table.Columns {
			if column.Null && (i > 0) {
				includeSQL = true
			}
		}
	}
	if includeSQL {
		s += "import\t\"database/sql\"\n"
	}
	return s
}

func (sm *StateMachine) ImportTime(ignoreNull bool) string {

	var s string
	var includeTime bool

	includeTime = false
	for _, table := range sm.Tables {
		for _, column := range table.Columns {
			switch column.Type {
			case "DATE", "TIME", "TIMESTAMP", "TIMESTAMPTZ", "INTERVAL":
				if ignoreNull || (!ignoreNull && !column.Null) {
					includeTime = true
				}
			default:
			}
		}
	}
	if includeTime {
		s += "import\t\"time\"\n"
	}
	return s
}

func (sm *StateMachine) ProjName() string {
	wd, _ := os.Getwd()
	return (filepath.Base(wd))
}

func (sm *StateMachine) DataPath() string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	gopath = filepath.Join(gopath, "src") + "/"
	wd, _ := os.Getwd()
	projpath := filepath.Join(strings.TrimPrefix(wd, gopath), "data")
	return projpath
}

// new
func (table Table) PlurName() string {
	return plural.Plural(lowerize(table.Name))
}

func (table Table) SingName() string {
	return plural.Singular(lowerize(table.Name))
}

// rename CapName => CapPlurName
func (table Table) CapPlurName() string {
	return capitalize(plural.Plural(lowerize(table.Name)))
}

func (table Table) CapSingName() string {
	return capitalize(plural.Singular(lowerize(table.Name)))
}

func (table Table) DropTableStatement() string {
	var s string
	s += "(\"DROP TABLE IF EXISTS " + table.Name + " CASCADE\")"
	return s
}

func (table Table) CreateTableStatement() string {
	var s string
	s += "(`" + table.Query + "`)"
	return s
}

// {<go type>, <go nulltype> <go nulltype field>}
var typeMap = map[string][]string{
	"BOOLEAN":     {"bool", "sql.NullBool", ".Bool"},
	"BOOL":        {"bool", "sql.NullBool", ".Bool"},
	"CHARID":      {"string", "sql.NullString", ".String"},
	"VARCHARID":   {"string", "sql.NullString", ".String"},
	"TEXT":        {"string", "sql.NullString", ".String"},
	"SMALLINT":    {"int16", "sql.NullInt32", ".Int32"},
	"INT":         {"int32", "sql.NullInt32", ".Int32"},
	"INTEGER":     {"int32", "sql.NullInt32", ".Int32"},
	"BIGINT":      {"int64", "sql.NullInt64", ".Int64"},
	"SMALLSERIAL": {"int16", "sql.NullInt32", ".Int32"},
	"SERIAL":      {"int32", "sql.NullInt32", ".Int32"},
	"BIGSERIAL":   {"int64", "sql.NullInt64", ".Int64"},
	"FLOATID":     {"float64", "sql.NullFloat64", ".Float64"},
	"REAL":        {"float32", "sql.NullFloat64", ".Float64"},
	"FLOAT8":      {"float32", "sql.NullFloat64", ".Float64"},
	"DECIMAL":     {"float64", "sql.NullFloat64", ".Float64"},
	"NUMERIC":     {"float64", "sql.NullFloat64", ".Float64"},
	"NUMERICID":   {"float64", "sql.NullFloat64", ".Float64"},
	"PRECISION":   {"float64", "sql.NullFloat64", ".Float64"}, //DOUBLE PRECISION
	"DATE":        {"time.Time", "sql.NullTime", ".Time"},
	"TIME":        {"time.Time", "sql.NullTime", ".Time"},
	"TIMESTAMPTZ": {"time.Time", "sql.NullTime", ".Time"},
	"TIMESTAMP":   {"time.Time", "sql.NullTime", ".Time"},
	"INTERVAL":    {"string", "sql.NullString", ".String"},
	"JSON":        {"string", "sql.NullString", ".String"},
	"JSONB":       {"string", "sql.NullString", ".String"},
	"UUID":        {"string", "sql.NullString", ".String"},
	"BYTEA":       {"[]byte", "[]byte", ""},
}

func (table Table) StructFields() string {

	var s string

	for i, column := range table.Columns {
		s += "\t" + camelize(column.Name) + " "
		if column.Null && (i > 0) {
			s += typeMap[column.Type][1]
		} else {
			s += typeMap[column.Type][0]
		}
		// s += " " + typeMap[column.Type][0]
		s += "`xml:\"" + camelize(column.Name) + "\" json:\"" + lowerize(camelize(column.Name)) + "\"`"
		s += "\n"
	}
	return s
}

func (table Table) Star() string {

	var s string
	for i, column := range table.Columns {
		s += " " + column.Name
		s += comma(i, len(table.Columns))
	}
	return s
}

func (table Table) ScanAll() string {

	var s string
	s += ".Scan("
	for i, column := range table.Columns {
		s += " &result." + camelize(column.Name)
		s += comma(i, len(table.Columns))
	}
	s += ")"
	return s
}

func (table Table) CreateStatement() string {
	var s string
	s += "(\"INSERT INTO " + table.Name + " ("

	for i, column := range table.Columns {
		if column.Generated {
			continue
		}
		s += " " + column.Name
		s += comma(i, len(table.Columns))
	}
	s += ") VALUES ("
	index := 1
	for i, column := range table.Columns {
		if column.Generated {
			continue
		}
		s += "$"
		s += strconv.Itoa(index)
		s += comma(i, len(table.Columns))
		index++
	}
	s += ") RETURNING"
	for i, column := range table.Columns {
		s += " " + column.Name
		s += comma(i, len(table.Columns))
	}
	s += "\")"
	return s
}

func (table Table) CreateQuery() string {

	var s string
	s += "("
	for i, column := range table.Columns {
		if column.Generated {
			continue
		}
		s += " " + table.SingName() + "." + camelize(column.Name)
		s += comma(i, len(table.Columns))
	}
	s += ")"
	s += table.ScanAll()
	return s
}

func (table Table) RetrieveStatement() string {
	var s string
	s += "(\"SELECT" + table.Star() + " FROM " + table.Name + " WHERE ("

	index := 1
	for i, column := range table.Columns {
		if i == 0 {
			s += column.Name + " = $" + strconv.Itoa(index)
			s += ")\", " + table.SingName() + "." + camelize(column.Name) + ")"
		}
		break
	}
	s += table.ScanAll()
	return s
}

func (table Table) RetrieveAllStatement() string {
	var s string
	s += "(\"SELECT" + table.Star() + " FROM " + table.Name + " ORDER BY "

	for i, column := range table.Columns {
		if i == 0 {
			s += column.Name
		}
		break
	}
	s += " ASC\")"
	return s
}

func (table Table) UpdateStatement() string {
	var s string
	s += "(\"UPDATE " + table.Name + " SET"

	index := 2
	for i, column := range table.Columns {
		if i == 0 {
			continue
		}
		s += " " + column.Name + " = $" + strconv.Itoa(index)
		index++
		s += comma(i, len(table.Columns))
	}
	s += " WHERE ("
	index = 1
	for i, column := range table.Columns {
		if i == 0 {
			s += column.Name + " = $" + strconv.Itoa(index)
			s += ") RETURNING"
		}
		break
	}
	for i, column := range table.Columns {
		s += " " + column.Name
		s += comma(i, len(table.Columns))
	}
	s += "\")"
	return s
}

func (table Table) UpdateQuery() string {

	var s string
	s += "("
	for i, column := range table.Columns {
		s += " " + table.SingName() + "." + camelize(column.Name)
		s += comma(i, len(table.Columns))
	}
	s += ")"
	s += table.ScanAll()
	return s
}

func (table Table) DeleteStatement() string {
	var s string
	s += "(\"DELETE FROM " + table.Name + " WHERE ("

	index := 1
	for i, column := range table.Columns {
		if i == 0 {
			s += column.Name + " = $" + strconv.Itoa(index)
		}
		break
	}
	s += ")\")"
	return s
}

func (table Table) DeleteQuery() string {

	var s string
	for i, column := range table.Columns {
		if i == 0 {
			s += "(" + table.SingName() + "." + camelize(column.Name) + ")"
		}
		break
	}
	return s
}

func (table Table) DeleteAllStatement() string {
	var s string
	s += "(\"DELETE FROM " + table.Name + "\")"
	return s
}

//TEST SPECIFIC

type GenerateFunc func(int, int, Column) string

type testFuncs struct {
	GenerateData GenerateFunc
	CompareData  string
}

var dataMap = map[string]testFuncs{
	"BOOLEAN":     {boolTestData, "defaultCompare"},
	"BOOL":        {boolTestData, "defaultCompare"},
	"CHARID":      {stringTestData, "defaultCompare"}, //Needs custom compare to work for all cases
	"VARCHARID":   {stringTestData, "defaultCompare"},
	"TEXT":        {stringTestData, "defaultCompare"},
	"SMALLINT":    {int16TestData, "defaultCompare"},
	"INT":         {int32TestData, "defaultCompare"},
	"INTEGER":     {int32TestData, "defaultCompare"},
	"BIGINT":      {int32TestData, "defaultCompare"},
	"SMALLSERIAL": {smallserialTestData, "defaultCompare"},
	"SERIAL":      {serialTestData, "defaultCompare"},
	"BIGSERIAL":   {bigserialTestData, "defaultCompare"},
	"FLOATID":     {float64TestData, "defaultCompare"}, //Needs custom compare to work for all cases
	"REAL":        {float32Trunc6TestData, "realCompare"},
	"FLOAT8":      {float32TestData, "defaultCompare"},
	"DECIMAL":     {float64TestData, "defaultCompare"},
	"NUMERIC":     {float64TestData, "defaultCompare"},
	"NUMERICID":   {float64TestData, "defaultCompare"}, //Needs custom compare to work for all cases
	"PRECISION":   {float64TestData, "defaultCompare"}, //DOUBLE PRECISION
	"DATE":        {dateTestData, "stringCompare"},     //use string compare to get around pq date issue
	"TIME":        {timeTestData, "stringCompare"},
	"TIMESTAMPTZ": {timestampTestData, "stringCompare"},
	"TIMESTAMP":   {timestampTestData, "stringCompare"},
	"INTERVAL":    {durationTestData, "defaultCompare"},
	"JSON":        {jsonTestData, "jsonCompare"}, //use jsonCompare since keys can be any order
	"JSONB":       {jsonbTestData, "jsonCompare"},
	"UUID":        {uuidTestData, "defaultCompare"},
	"BYTEA":       {byteaTestData, "byteaCompare"},
}

func (table Table) CompareMapFields() string {
	var s string

	for _, column := range table.Columns {
		s += "\t\"" + camelize(column.Name) + "\": "
		s += dataMap[column.Type].CompareData + ",\n"
	}
	return s
}

func (table Table) TestData(dataid int) string {
	rand.Seed(time.Now().UnixNano())

	var s string

	s = ""
	for columnid, column := range table.Columns {
		s += " " + camelize(column.Name) + ": " + dataMap[column.Type].GenerateData(dataid, columnid, column)
		s += comma(columnid, len(table.Columns))
	}
	return s
}

func nullPrefix(datatype string, column Column) string {
	if column.Null {
		return (typeMap[datatype][1] + "{")
	} else {
		return ""
	}

}

func nullSuffix(column Column) string {
	if column.Null {
		return (", true}")
	} else {
		return ""
	}
}

func boolTestData(dataid int, columnid int, column Column) string {
	s := nullPrefix("BOOL", column)
	s += strconv.FormatBool(rand.Intn(2) != 0)
	s += nullSuffix(column)
	return (s)
}
func stringTestData(dataid int, columnid int, column Column) string {
	s := nullPrefix("TEXT", column)
	s += "\"" + randString(16) + "\""
	s += nullSuffix(column)
	return (s)
}
func int16TestData(dataid int, columnid int, column Column) string {
	s := ""
	if columnid == 0 { //assume primary
		s += strconv.FormatInt(int64(dataid), 10)
	} else if column.Ref {
		s += nullPrefix("INT", column)
		s += strconv.FormatInt(int64(dataid), 10)
		s += nullSuffix(column)
	} else {
		s += nullPrefix("INT", column)
		s += strconv.FormatInt(int64(rand.Intn(32767)), 10)
		s += nullSuffix(column)
	}
	return s
}
func int32TestData(dataid int, columnid int, column Column) string {
	s := ""
	if columnid == 0 { //assume primary
		s += strconv.FormatInt(int64(dataid), 10)
	} else if column.Ref {
		s += nullPrefix("INT", column)
		s += strconv.FormatInt(int64(dataid), 10)
		s += nullSuffix(column)
	} else {
		s += nullPrefix("INT", column)
		s += strconv.FormatInt(int64(rand.Int31()), 10)
		s += nullSuffix(column)
	}
	return s
}
func int64TestData(dataid int, columnid int, column Column) string {
	s := ""
	if columnid == 0 { //assume primary
		s += strconv.FormatInt(int64(dataid), 10)
	} else if column.Ref {
		s += nullPrefix("BIGINT", column)
		s += strconv.FormatInt(int64(dataid), 10)
		s += nullSuffix(column)
	} else {
		s += nullPrefix("BIGINT", column)
		s += strconv.FormatInt(rand.Int63(), 10)
		s += nullSuffix(column)
	}
	return s
}
func smallserialTestData(dataid int, columnid int, column Column) string {
	s := ""
	if columnid == 0 { //assume primary
		s += strconv.FormatInt(int64(dataid), 10)
	} else if column.Ref {
		s += nullPrefix("SMALLSERIAL", column)
		s += strconv.FormatInt(int64(dataid), 10)
		s += nullSuffix(column)
	} else {
		s += nullPrefix("SMALLSERIAL", column)
		s += strconv.FormatInt(int64(dataid), 10)
		s += nullSuffix(column)
	}
	return s
}
func serialTestData(dataid int, columnid int, column Column) string {
	s := ""
	if columnid == 0 { //assume primary
		s += strconv.FormatInt(int64(dataid), 10)
	} else if column.Ref {
		s += nullPrefix("SERIAL", column)
		s += strconv.FormatInt(int64(dataid), 10)
		s += nullSuffix(column)
	} else {
		s += nullPrefix("SERIAL", column)
		s += strconv.FormatInt(int64(dataid), 10)
		s += nullSuffix(column)
	}
	return s
}
func bigserialTestData(dataid int, columnid int, column Column) string {
	s := ""
	if columnid == 0 { //assume primary
		s += strconv.FormatInt(int64(dataid), 10)
	} else if column.Ref {
		s += nullPrefix("BIGSERIAL", column)
		s += strconv.FormatInt(int64(dataid), 10)
		s += nullSuffix(column)
	} else {
		s += nullPrefix("BIGSERIAL", column)
		s += strconv.FormatInt(int64(dataid), 10)
		s += nullSuffix(column)
	}
	return s
}
func float64TestData(dataid int, columnid int, column Column) string {
	s := nullPrefix("DECIMAL", column)
	s += strconv.FormatFloat(rand.NormFloat64(), 'f', -1, 64)
	s += nullSuffix(column)
	return s
}
func float32TestData(dataid int, columnid int, column Column) string {
	s := nullPrefix("REAL", column)
	s += strconv.FormatFloat(float64(rand.Float32()), 'f', -1, 32)
	s += nullSuffix(column)
	return s
}
func float32Trunc6TestData(dataid int, columnid int, column Column) string {
	s := nullPrefix("REAL", column)
	s += strconv.FormatFloat(float64(rand.Float32()), 'f', 6, 32)
	s += nullSuffix(column)
	return s
}
func dateTestData(dataid int, columnid int, column Column) string {
	s := nullPrefix("DATE", column)
	s += "time.Now().UTC().Truncate(time.Hour * 24)"
	s += nullSuffix(column)
	return s
}
func timeTestData(dataid int, columnid int, column Column) string {
	s := nullPrefix("TIME", column)
	s += "time.Date(0000, time.January, 1, time.Now().UTC().Hour(), time.Now().UTC().Minute(), time.Now().UTC().Second(), time.Now().UTC().Nanosecond(), time.UTC).Truncate(time.Microsecond)"
	s += nullSuffix(column)
	return s
}
func timestampTestData(dataid int, columnid int, column Column) string {
	s := nullPrefix("TIMESTAMPTZ", column)
	s += "time.Now().UTC().Truncate(time.Microsecond)"
	s += nullSuffix(column)
	return s
}
func durationTestData(dataid int, columnid int, column Column) string {
	s := nullPrefix("INTERVAL", column)
	s += "\"12:34:45\""
	s += nullSuffix(column)
	return s
}
func jsonTestData(dataid int, columnid int, column Column) string {
	s := nullPrefix("JSON", column)
	s += randJson()
	s += nullSuffix(column)
	return s
}
func jsonbTestData(dataid int, columnid int, column Column) string {
	s := nullPrefix("JSONB", column)
	s += randJson()
	s += nullSuffix(column)
	return s
}
func uuidTestData(dataid int, columnid int, column Column) string {
	s := nullPrefix("UUID", column)
	s += "\"" + randUUID() + "\""
	s += nullSuffix(column)
	return s
}

func byteaTestData(dataid int, columnid int, column Column) string {
	s := ""
	if columnid == 0 || column.Ref {
		s += "[]byte{0,0,0,0,"
		s += strconv.FormatInt(int64(dataid), 10)
		s += "}"
	} else {
		s += randBytea(20)
	}
	return s
}

// []byte{0,0,0,1}
func randBytea(length int) string {
	str := "[]byte{"

	for i := 0; i < length; i++ {
		str += strconv.FormatInt(int64(rand.Intn(256)), 10)
		if i < length-1 {
			str += ","
		}
	}
	str += "}"
	return str
}

func randString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func randJson() string {
	return "\"{\\\"name\\\": \\\"" + randString(16) + "\\\", \\\"age\\\": " + strconv.FormatInt(int64(rand.Int31()), 10) + ", \\\"city\\\": \\\"" + randString(20) + "\\\"}\""
}

func randUUID() (uuid string) {
	u := new([16]byte)
	_, err := rand.Read(u[:])
	if err != nil {
		log.Panicln("Cannot generate UUID", err.Error())
	}
	u[8] = (u[8] | 0x40) & 0x7F
	u[6] = (u[6] & 0xF) | (0x4 << 4)
	uuid = fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
	return
}
