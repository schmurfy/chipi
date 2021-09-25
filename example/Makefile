
run: build
	./example

build: generate
	go build -o example .

generate: chipi-gen
	./chipi-gen -dir .

chipi-gen:
	go build -o chipi-gen ../chipi-gen/gen.go
