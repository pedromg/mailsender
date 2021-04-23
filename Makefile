testall:
	go test -v ./...

testrace:
	go test -race ./...

build: testall
	go build -o mailsender mailsender.go

buildlinux: testall
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o mailsender.linux mailsender.go
