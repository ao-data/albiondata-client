.PHONY: run frontend fmt validate-fmt build-windows build-linux build-darwin

run:
	scripts/run.sh

frontend:
	cd frontend && npm ci && npm run build

fmt:
	scripts/fmt.sh

validate-fmt:
	scripts/validate-fmt.sh

build-windows:
	scripts/build-windows.sh

build-linux:
	scripts/build-linux.sh

build-darwin:
	scripts/build-darwin.sh
