default: static

static:
	go build -tags "netgo" -ldflags "-w -linkmode external -extldflags -static" -v -o fumble
