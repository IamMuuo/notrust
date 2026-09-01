# not rust. just go, just build, just done.

bin_dir := "bin"
daemon  := "notrusted"
cli     := "notrust"

default:
    just --list

# build both binaries into ./bin
build: build-daemon build-cli

build-daemon:
    go build -o {{bin_dir}}/{{daemon}} ./cmd/notrustd

build-cli:
    go build -o {{bin_dir}}/{{cli}} ./cmd/notrust

# run the daemon in the foreground with debug logging, using local config
dev: build-daemon
    {{bin_dir}}/{{daemon}} --log-level=debug --config=./config.example.yaml

# quick status check against a running daemon
status: build-cli
    {{bin_dir}}/{{cli}} status

# install both binaries to $GOPATH/bin
install:
    go install ./cmd/notrustd
    go install ./cmd/notrust

# fast unit tests only, no real docker daemon needed
test:
    go test ./...

# same, with the race detector on, slower but worth running before pushing
test-race:
    go test -race ./...

# integration tests, needs a real docker daemon
test-integration:
    go test -tags=integration ./test/integration/...

fmt:
    gofmt -l -w .
    goimports-reviser -format .

vet:
    go vet ./...

lint:
    golangci-lint run ./...

tidy:
    go mod tidy

clean:
    rm -rf {{bin_dir}}

# fmt, vet, lint, unit tests, in that order, stop on first failure
check: fmt vet lint test

# cross compile release binaries for linux
release:
    GOOS=linux GOARCH=amd64 go build -o {{bin_dir}}/{{daemon}}-linux-amd64 ./cmd/notrustd
    GOOS=linux GOARCH=arm64 go build -o {{bin_dir}}/{{daemon}}-linux-arm64 ./cmd/notrustd

# install as a per user systemd service, needed so notify-send/D-Bus works
install-service: build-daemon
    mkdir -p ~/.local/bin ~/.config/systemd/user ~/.config/notrust
    cp {{bin_dir}}/{{daemon}} ~/.local/bin/{{daemon}}
    cp notrust.service ~/.config/systemd/user/notrust.service
    cp -n config.example.yaml ~/.config/notrust/config.yaml
    systemctl --user daemon-reload
    systemctl --user enable --now notrust.service

# tail daemon logs
logs:
    journalctl --user -u notrust -f
