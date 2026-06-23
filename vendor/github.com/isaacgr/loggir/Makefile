.PHONY: build test tests bench maintainer-clean

maintainer-clean: clean

tests: test bench

test:
	go test -race -cover -bench=^$$ ./...

bench:
	go test -run=^$$ -bench=. -benchmem -count=5 -cpu=1,8,16,24 ./...
