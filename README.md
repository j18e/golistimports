# golistimports

Golistimports is a commandline tool which reports on the packages imported in a
Go project. Simply navigate to the project root and run `golistimports`, and it
will give you an overview of:
1. builtin packages which are imported by the project
2. packages in the same module which are imported by the project
3. non-builtin packages which are imported by the project
4. packages imported by non-builtin packages imported by the project (go.sum)

## Installing
`go get -u github.com/j18e/golistimports`
