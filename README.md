# metaapi

Metaprogramming example to generate go CRUD api from sql table definitions

See original Medium article here. (tbd)

To run:

* Clone the metaapi project and compile
* Create a directory for your new project
* Copy the crud template (crud.txt), your sql table definitions (todo.sql), and optionally a file to kick off go generate (todo.go)
* Run go generate or just manually run metaapi
* Your new go file will be created as _generated.go (based on the name of your sql file)

Example:

```
go get github.com/exyzzy/metaapi
go install $GOPATH/src/github.com/exyzzy/metaapi
mkdir myproj
cd myproj
cp $GOPATH/src/github.com/exyzzy/metaapi/metasql/crud.txt .
cp $GOPATH/src/github.com/exyzzy/metaapi/metasql/todo.sql .
cp $GOPATH/src/github.com/exyzzy/metaapi/metasql/todo.go .
go generate
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
SQL	        GO
-----------|-------
BOOLEAN	    bool
BOOL	    bool
CHAR(n)	    string
VARCHAR(n)	string
TEXT	    string
SMALLINT	int16
INT	        int32
INTEGER	    int32
BIGINT	    int64
SMALLSERIAL	int16
SERIAL	    int32
BIGSERIAL	int64
FLOAT(n)	float64
REAL	    float32
FLOAT8	    float32
DECIMAL	    float64
NUMERIC	    float64
NUMERIC(p,s)	    float64
DOUBLE PRECISION	float64
DATE	    time.Time
TIME	    time.Time
TIMESTAMPTZ	time.Time
TIMESTAMP	time.Time
INTERVAL	time.Time
JSON	    []byte
JSONB	    []byte
UUID	    string
```

## State Machine
```
  CurState    Input  NextState     Action
  ---------------------------------------------
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