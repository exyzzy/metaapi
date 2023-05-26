# << .ProjName >>

Create

```
go get github.com/exyzzy/metaproj
go get github.com/exyzzy/metaapi
go install $GOPATH/src/github.com/exyzzy/metaproj
go install $GOPATH/src/github.com/exyzzy/metaapi
metaproj -proj=<< .ProjName >> -sql=<< .FName >> -type=vue
cd << .ProjName >>
go generate
```

Test, & Run:

```
createuser -P -d << .ProjName >> <pass: << .ProjName >> >
createdb << .ProjName >>
cd data
go test
cd ..
go test
go install
<< .ProjName >>
```


# API

```
<< range $index, $table := .Tables >>
DELETE /api/<< $table.SingName >>/createtable
POST /api/<< $table.SingName >>
GET /api/<< $table.SingName >>
GET /api/<< $table.SingName >>
PUT /api/<< $table.SingName >>
DELETE /api/<< $table.SingName >>
DELETE /api/<< $table.PlurName >>

type << $table.CapSingName >> struct {
<< $table.StructFields >>
}

<< end >>
```