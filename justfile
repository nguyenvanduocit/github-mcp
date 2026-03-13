build:
  CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/github-mcp ./main.go

build-cli:
  CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/github-cli ./cmd/cli/

dev:
  go run main.go --env .env --sse_port 3000

install:
  go install ./...

install-cli:
  go install ./cmd/cli/
