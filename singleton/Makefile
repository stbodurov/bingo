.PHONY: run test wasm build clean

run: wasm
	go run ./cmd/singleton

test:
	go test ./...

wasm:
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -trimpath -o web/static/wasm/wallet.wasm ./cmd/wallet-wasm
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" web/static/wasm/ 2>/dev/null || cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" web/static/wasm/

build: wasm
	go build -o bin/singleton ./cmd/singleton

clean:
	rm -rf bin data web/static/wasm/wallet.wasm
