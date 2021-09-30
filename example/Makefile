
run: build
	./example

build: generate
	go build -o example .

generate: chipi-gen
	./chipi-gen -dir .

.PHONY: chipi-gen
chipi-gen:
	go build -o chipi-gen ../chipi-gen/gen.go
