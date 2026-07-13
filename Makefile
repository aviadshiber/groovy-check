.PHONY: build install

build:
	go build -o groovy-check .

install: build
	mkdir -p $(HOME)/.local/bin
	cp groovy-check $(HOME)/.local/bin/groovy-check
	@echo "installed to $(HOME)/.local/bin/groovy-check — ensure it's on PATH"
