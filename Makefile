test:
	rm -f testdata/*.sqlite
	go test

huge:
	HUGE=true ${MAKE} test

tidy:
	go mod tidy

bench:
	go test -bench . -benchmem
