.PHONY: tests test bench clean

tests: test bench

test:
	-go test -race -cover -coverprofile=coverage.out -bench=^$$ ./...
	go tool cover -html=coverage.out -o coverage.html

bench:
	go test -run=^$$ -bench=. -benchmem -count=5 -cpu=1,8,16 ./...

clean:
	rm coverage.out coverage.html
