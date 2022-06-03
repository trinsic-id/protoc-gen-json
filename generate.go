package main

//go:generate go build -o ./protoc-gen-json.exe main.go option_parsing.go context.go util.go
//go:generate pwsh -Command "copy ./protoc-gen-json.exe C:/bin/protoc-gen-json.exe"
//go:generate protoc  --plugin=protoc-gen-json=./protoc-gen-json.exe -I. --json_out=./ ./*.proto
