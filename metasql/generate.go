package metasql

import (
	"errors"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"text/template"
)

//Generate assumes that the primary ID is in the first column (index 0)
func Generate(sm *StateMachine, txtFile string) error {
	if sm.FName == "" {
		return (errors.New("No file name"))
	}
	dot := strings.Index(sm.FName, ".")
	var prefix string
	if dot > 0 {
		prefix = sm.FName[:dot]
	} else {
		prefix = sm.FName
	}
	dat, err := ioutil.ReadFile("./" + txtFile)
	if err != nil {
		return err
	}

	tt := template.Must(template.New(prefix).Parse(string(dat)))
	dest := prefix + "_generated.go"
	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	tt.Execute(file, sm)
	file.Close()
	return nil
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
	s += "(\"DROP TABLE IF EXISTS " + table.Name + "\")"
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
		"INTERVAL":    "time.Time",
		"JSON":        "[]byte",
		"JSONB":       "[]byte",
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
