pre-build:
	go get github.com/shurcooL/graphql gopkg.in/yaml.v2
build:
	go build -tags "osusergo netgo" -ldflags="-extldflags=-static -s -w" scrape.go
