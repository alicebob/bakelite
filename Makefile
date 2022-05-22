test:
	rm -f testdata/*.sqlite
	go test

tidy:
	go mod tidy
