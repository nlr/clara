build:
	@go build -o bin/clara.exe

run: build
	@./bin/clara.exe