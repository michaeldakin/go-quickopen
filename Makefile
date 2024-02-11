build:
	go build -o bin/go-quickopen

run: build
	./bin/go-quickopen

prod: build
	cp ./bin/go-quickopen ~/.local/bin
	chmod +x ~/.local/bin/go-quickopen
