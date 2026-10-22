.PHONY: dev mac-dev watch build build-mac test test-integration generate mock

dev:
	cd app && go run ./cmd/mrav

mac-dev:
	cd app && go run -exec ./dev-sign.sh ./cmd/mrav

ifeq ($(OS),Windows_NT)
AIR_CONFIG = .air.windows.toml
else
AIR_CONFIG = .air.toml
endif
watch:
	cd app && air -c $(AIR_CONFIG)

build:
	cd app && go build -o bin/mrav ./cmd/mrav

build-mac:
	app/build-mac.sh

test:
	cd app && ginkgo run -r --randomize-all --label-filter='!integration'

test-integration:
	cd app && ginkgo run -r --randomize-all --label-filter='integration'

generate:
	cd orm/ent && go generate ./...

mock:
	cd app && go run github.com/vektra/mockery/v3@v3.8.1
