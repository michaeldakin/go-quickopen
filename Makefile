build:
	go build -o bin/go-quickopen

run: build
	./bin/go-quickopen

buildcp: build
	cp ./bin/go-quickopen ~/.local/bin
	ls -latr ~/.local/bin/go-quickopen
