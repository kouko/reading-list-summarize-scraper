.PHONY: build js js-quick clean generate

build: js
	go build -ldflags="-s -w" -o rlss ./cmd/rlss/

# 強制重新安裝 defuddle + 打包（預設）
js:
	rm -rf node_modules
	npm install
	npm run build:defuddle
	@echo "Bundled defuddle $$(npm list defuddle --depth=0 | grep defuddle)"

# 快速打包（不重裝，用於開發）
js-quick: node_modules
	npm run build:defuddle

node_modules: package.json package-lock.json
	npm install
	@touch node_modules

generate: node_modules
	go generate ./...

clean:
	rm -f rlss embed/defuddle.min.js
	rm -rf node_modules
