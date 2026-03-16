default: build

build:
		go build -o dink-filter ./cmd/filter/

clean:
		rm dink-filter
