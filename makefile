.PHONY: dev start build

dev: 
	air

start:
	go run main.go

build:
	go build -o dist/argus.exe
	cp argus-config.yml dist/argus-config.yml

clean:
	rm -Rf dist/argus.exe
	rmdir dist