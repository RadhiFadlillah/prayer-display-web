JS_DIR = internal/view/js
CSS_DIR = internal/view/css
LESS_DIR = internal/view/less

BROWSER_LIST = last 2 versions, > 8%
LESS_FLAGS = --strict-imports --clean-css --autoprefix="$(BROWSER_LIST)"
ESBUILD_FLAGS = --bundle

dev: less-bundle
	@go generate
	@go build -o prayer-display-web-dev

prod: js-bundle less-bundle
	@go generate
	@go build -tags prod

js-bundle:
	@npx esbuild --bundle --outdir="$(JS_DIR)" "$(JS_DIR)/entries/index.js"
	@npx esbuild --bundle --minify --outdir="$(JS_DIR)" "$(JS_DIR)/entries/index.js"

less-bundle:
	@npx lessc $(LESS_FLAGS) "$(LESS_DIR)/index.less" "$(CSS_DIR)/index.css"

install-npm:
	@npm install esbuild less less-plugin-autoprefix less-plugin-clean-css --save-dev