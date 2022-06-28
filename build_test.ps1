Set-Location $PSScriptRoot
go build
protoc --plugin="protoc-gen-json=${PSScriptRoot}/protoc-gen-json.exe" --json_out="./" --json_opt="test.json" ./test.proto

