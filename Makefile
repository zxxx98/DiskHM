test:
	go test ./...

build:
	go build ./cmd/diskhmd

ui-test:
	npm --prefix web run test -- --run

ui-build:
	npm --prefix web run build
