ENV      ?= "development"
APP_HOST ?= "localhost"
APP_PORT ?= "9000"

run:
	export ENV=$(ENV) \
         APP_HOST=$(APP_HOST) \
         APP_PORT=$(APP_PORT) \
         go clean; \
         go build; \
         ./api-gateway
