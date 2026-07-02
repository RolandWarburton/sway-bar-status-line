VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
go build -ldflags "-X main.version=${VERSION}" -o statusBar
