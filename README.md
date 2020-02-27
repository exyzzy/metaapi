# metaapi (v2.0)

Metaprogramming example to generate go CRUD api and test functions from sql table definitions

Pull the v1.0 tag metaapi commit to have code that matches the original Medium article: 
* https://levelup.gitconnected.com/metaprogram-in-go-5a2a7e989613


Current v2.0 code matches this article:
* (tbd, in progress)

See also:
* http://github.com/exyzzy/metaproj
* http://github.com/exyzzy/pipe

Example Using Internal Files, no metaproj

```
go get github.com/exyzzy/metaapi
go install $GOPATH/src/github.com/exyzzy/metaapi
mkdir myproj
cd myproj
cp $GOPATH/src/github.com/exyzzy/metaapi/examples/todo.sql .
cp $GOPATH/src/github.com/exyzzy/metaapi/examples/todo.go .
go generate
```
Example Using Internal Files, with metaproj

```
go get github.com/exyzzy/metaapi
go install $GOPATH/src/github.com/exyzzy/metaapi
go get github.com/exyzzy/metaproj
go install $GOPATH/src/github.com/exyzzy/metaproj
cp $GOPATH/src/github.com/exyzzy/metaapi/examples/alltypes.sql .
metaproj -sql=alltypes.sql -proj=myproj 
cd myproj
createuser -P -d myproj
# use password: myproj
createdb myproj
go generate
go test
```

Example Using External Files, with metaproj and pipe

```
go get github.com/exyzzy/metaapi
go install $GOPATH/src/github.com/exyzzy/metaapi
go get github.com/exyzzy/metaproj
go install $GOPATH/src/github.com/exyzzy/metaproj
go get github.com/exyzzy/pipe
go install $GOPATH/src/github.com/exyzzy/pipe
cp $GOPATH/src/github.com/exyzzy/metaapi/examples/alltypes.sql .
metaproj -sql=alltypes.sql -proj=myproj -type=external 
cd myproj
createuser -P -d myproj
# use password: myproj
createdb myproj
go install
go generate
go test
```


## Grammar
```
CREATE [some_stuff]* TABLE [IF NOT EXISTS] table_name (
    column_name data_type [some_stuff]* 
    [, ...]
) [some_stuff]* ;
```

## Supported SQL Data Types and Conversion to Go
```
SQL	           GO
--------------|-------
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
```

## State Machine
```
 CurState  Input  NextState   Action
 ------------------------------------------
 0    Error          0     error_state     
 1    CREATE         2     create_table     
 2    TABLE          3     nop     
 2    ID             2     some_stuff     
 3    IF             4     nop     
 4    NOT            5     nop     
 5    EXISTS         3     nop     
 3    ID             6     table_name     
 6    (              7     nop     
 7    ID             8     column_name     
 7    UUID           8     column_name     
 8    BOOLEAN        9     data_type     
 8    BOOL           9     data_type     
 8    CHARID         9     data_type     
 8    VARCHARID      9     data_type     
 8    TEXT           9     data_type     
 8    SMALLINT       9     data_type     
 8    INT            9     data_type     
 8    INTEGER        9     data_type     
 8    BIGINT         9     data_type     
 8    SMALLSERIAL    9     data_type     
 8    SERIAL         9     data_type     
 8    BIGSERIAL      9     data_type     
 8    FLOATID        9     data_type     
 8    REAL           9     data_type     
 8    FLOAT8         9     data_type     
 8    DECIMAL        9     data_type     
 8    NUMERIC        9     data_type     
 8    NUMERICID      9     data_type     
 8    DOUBLE         10    nop     
 10   PRECISION      9     data_type     
 8    DATE           9     data_type     
 8    TIME           9     data_type     
 8    TIMESTAMPTZ    9     data_type     
 8    TIMESTAMP      9     data_type     
 8    INTERVAL       9     data_type     
 8    JSON           9     data_type     
 8    JSONB          9     data_type     
 8    UUID           9     data_type     
 9    ,              7     nop     
 9    )              11    nop     
 9    REFID          9     some_stuff     
 9    NOT            9     some_stuff     
 9    ID             9     some_stuff     
 11   ;              1     end_table     
 11   ID             11    some_stuff     
```
