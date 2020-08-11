$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o .\dist\main main.go
build-lambda-zip.exe -output .\dist\main.zip .\dist\main