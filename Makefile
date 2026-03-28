.PHONY: build js clean update-defuddle generate

build: js
	go build -ldflags="-s -w" -o rlss ./cmd/rlss/

js: node_modules
	npm run build:defuddle

node_modules: package.json package-lock.json
	npm install
	@touch node_modules

update-defuddle:
	npm update defuddle
	npm run build:defuddle
	@echo "Updated defuddle to $$(npm list defuddle --depth=0 | grep defuddle)"

generate: node_modules
	go generate ./...

clean:
	rm -f rlss embed/defuddle.min.js
	rm -rf node_modules
