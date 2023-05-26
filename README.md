# metaapi (v4.0)

New for v4:
* Support BYTEA
* Support GENERATED

Generate go CRUD and HTTP api and test functions from sql table definitions

New for v3:
* Project setup for a fully functioning web server with Vue/Vuetify
* Automatic public HTTP api (routes and handlers, Gorilla Mux) on top of the CRUD api
* Automatic tests for the HTTP api
* Support for Go/PostgreSQL NULL fields/data
* Simple Vuetify table view page for data

v3.0 matches Medium article:
* Automatic Applications in Go

## Previous versions:

Pull the v1.0 tag metaapi commit to have code that matches the original Medium article: 
* https://levelup.gitconnected.com/metaprogram-in-go-5a2a7e989613


Pull the v2.0 tag metaapi commit to have code that matches the Medium article:
* https://levelup.gitconnected.com/automatic-testing-in-go-ce581238eb57

See also:
* http://github.com/exyzzy/metaproj
* http://github.com/exyzzy/metasplice
* http://github.com/exyzzy/pipe

## Most Common Scenario:

```
#assume project and database name: todo (can be anything)
#assume sql file: events.sql (from examples, but can be anything)
createuser -P -d todo <pass: todo>
createdb todo
go get github.com/exyzzy/metaapi
go install $GOPATH/src/github.com/exyzzy/metaapi
go get github.com/exyzzy/metaproj
go install $GOPATH/src/github.com/exyzzy/metaproj
cp $GOPATH/src/github.com/exyzzy/metaapi/examples/events.sql .
metaproj -sql=events.sql -proj=todo -type=vue
cd todo
go generate
go install
go test
```

## Legacy:

## Example Using Internal Files, no metaproj

```
go get github.com/exyzzy/metaapi
go install $GOPATH/src/github.com/exyzzy/metaapi
mkdir myproj
cd myproj
cp $GOPATH/src/github.com/exyzzy/metaapi/examples/todo.sql .
cp $GOPATH/src/github.com/exyzzy/metaapi/examples/todo.go .
go generate
```
## Example Using Internal Files, with metaproj

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

## Example Using External Files, with metaproj and pipe

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
