.PHONY: build clean deploy deploy_func remove

build:
	go mod tidy
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/main src/handler/main.go

clean:
	rm -rf ./bin

deploy: clean build
	npx sls deploy --verbose

deploy_func: clean build
	npx sls deploy function -f api --verbose

remove:
	npx sls remove --verbose