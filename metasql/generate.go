package metasql

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"reflect"
	"time"

	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/exyzzy/metaapi/data"
)

type Column struct {
	Name string
	Type string
}

type Table struct {
	Name    string
	Query   string
	Columns []Column
}

type StateMachine struct {
	FName    string
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

//Generate assumes that the primary ID is in the first column (index 0)
func Generate(dta interface{}, txtFile string) error {

	var sm StateMachine
	var prefix string
	var suffix string
	// fmt.Println("GENERATE: ", reflect.TypeOf(dta).Name())
	if reflect.TypeOf(dta).Name() == "StateMachine" {
		sm = dta.(StateMachine)

		if sm.FName == "" {
			return (errors.New("No file name"))
		}
		prefix = sm.FilePrefix()
	}
	if txtFile != "" {
		dot := strings.Index(txtFile, ".")
		if dot > 0 {
			suffix = txtFile[:dot]
		} else {
			suffix = txtFile
		}

		dest := prefix + "_generated_" + suffix + ".go"

		//for -pipe option, instead of data.Asset, use:
		// dat, err := ioutil.ReadFile("./" + txtFile)
		// if err != nil {
		// 	return err
		// }

		dat, err := data.Asset(txtFile)
		if err != nil {
			return err
		}
		return generateFile(dat, &sm, dest)
	}
	return nil
}

func generateFile(templatesrc []byte, data interface{}, dest string) error {
	tt := template.Must(template.New("file").Parse(string(templatesrc)))
	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	err = tt.Execute(file, data)
	file.Close()
	return err
}

//======== string helpers

//should use: https://github.com/blakeembrey/pluralize
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

//turns submitted_at into SubmittedAt, and otherwise capitalizes
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

// Writing it to be extended
func (sm *StateMachine) Import() string {

	var s string
	var includeTime bool

	includeTime = false
	for _, table := range sm.Tables {
		for _, column := range table.Columns {
			switch column.Type {
			case "DATE", "TIME", "TIMESTAMP", "TIMESTAMPTZ", "INTERVAL":
				includeTime = true
			default:
			}
		}
	}
	s += "import (\n\t\"database/sql\"\n"
	s += "\t_ \"github.com/lib/pq\"\n"

	if includeTime {
		s += "\t\"time\"\n"
	}
	s += ")"
	return s
}

func (table Table) SingName() string {
	return singularize(table.Name)
}

func (table Table) CapName() string {
	return capitalize(lowerize(table.Name))
}

func (table Table) CapSingName() string {
	return capitalize(singularize(table.Name))
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

func (table Table) StructFields() string {

	var typeMap = map[string]string{
		"BOOLEAN":     "bool",
		"BOOL":        "bool",
		"CHARID":      "string",
		"VARCHARID":   "string",
		"TEXT":        "string",
		"SMALLINT":    "int16",
		"INT":         "int32",
		"INTEGER":     "int32",
		"BIGINT":      "int64",
		"SMALLSERIAL": "int16",
		"SERIAL":      "int32",
		"BIGSERIAL":   "int64",
		"FLOATID":     "float64",
		"REAL":        "float32",
		"FLOAT8":      "float32",
		"DECIMAL":     "float64",
		"NUMERIC":     "float64",
		"NUMERICID":   "float64",
		"PRECISION":   "float64", //DOUBLE PRECISION
		"DATE":        "time.Time",
		"TIME":        "time.Time",
		"TIMESTAMPTZ": "time.Time",
		"TIMESTAMP":   "time.Time",
		"INTERVAL":    "string",
		"JSON":        "string",
		"JSONB":       "string",
		"UUID":        "string",
	}
	var s string

	for _, column := range table.Columns {
		s += "\t" + camelize(column.Name)
		s += " " + typeMap[column.Type]
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
		if i == 0 {
			continue
		}
		s += " " + column.Name
		s += comma(i, len(table.Columns))
	}
	s += ") VALUES ("
	index := 1
	for i, _ := range table.Columns {
		if i == 0 {
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
		if i == 0 {
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
	s += " DESC\")"
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

type GenerateFunc func(int, int) string

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
	"SMALLSERIAL": {serialTestData, "defaultCompare"},
	"SERIAL":      {serialTestData, "defaultCompare"},
	"BIGSERIAL":   {serialTestData, "defaultCompare"},
	"FLOATID":     {float64TestData, "defaultCompare"}, //Needs custom compare to work for all cases
	"REAL":        {float32TestData, "defaultCompare"},
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

	s = "{"
	for columnid, column := range table.Columns {
		s += " " + dataMap[column.Type].GenerateData(dataid, columnid)
		s += comma(columnid, len(table.Columns))
	}
	s += "}"
	return s
}

func boolTestData(dataid int, columnid int) string {
	return (strconv.FormatBool(rand.Intn(2) != 0))
}
func stringTestData(dataid int, columnid int) string {
	return ("\"" + randString(16) + "\"")
}
func int16TestData(dataid int, columnid int) string {
	if columnid == 0 { //assume serial
		return (strconv.FormatInt(int64(dataid), 10))
	} else {
		return (strconv.FormatInt(int64(rand.Intn(32767)), 10))
	}
}
func int32TestData(dataid int, columnid int) string {
	if columnid == 0 { //assume serial
		return (strconv.FormatInt(int64(dataid), 10))
	} else {
		return (strconv.FormatInt(int64(rand.Int31()), 10))
	}
}
func int64TestData(dataid int, columnid int) string {
	if columnid == 0 { //assume serial
		return (strconv.FormatInt(int64(dataid), 10))
	} else {
		return (strconv.FormatInt(rand.Int63(), 10))
	}
}
func serialTestData(dataid int, columnid int) string {
	return strconv.Itoa(dataid)
}
func float64TestData(dataid int, columnid int) string {
	return (strconv.FormatFloat(rand.NormFloat64(), 'f', -1, 64))
}
func float32TestData(dataid int, columnid int) string {
	return (strconv.FormatFloat(float64(rand.Float32()), 'f', -1, 32))
}
func timeTestData(dataid int, columnid int) string {
	return "time.Date(0000, time.January, 1, time.Now().UTC().Hour(), time.Now().UTC().Minute(), time.Now().UTC().Second(), time.Now().UTC().Nanosecond(), time.UTC)"
}
func timestampTestData(dataid int, columnid int) string {
	return "time.Now().UTC().Truncate(time.Microsecond)"
}
func durationTestData(dataid int, columnid int) string {
	return "\"12:34:45\""
}
func dateTestData(dataid int, columnid int) string {
	return "time.Now().UTC().Truncate(time.Hour * 24)"
}
func jsonTestData(dataid int, columnid int) string {
	return randJson()
}
func jsonbTestData(dataid int, columnid int) string {
	return randJson()
}
func uuidTestData(dataid int, columnid int) string {
	return "\"" + randUUID() + "\""
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
