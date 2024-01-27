build:
	@go build -o bin/clara

run: build
	@./bin/clara
