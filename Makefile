.PHONY: init
init:
	go install github.com/golang/mock/mockgen@latest
	go install github.com/swaggo/swag/v2/cmd/swag@v2.0.0-rc5

.PHONY: admin-bootstrap admin-run admin-migrate admin-web-build admin-build admin-build-task admin-build-migration admin-test admin-mock admin-docker-task admin-swag
admin-bootstrap:
	docker compose -f ./deploy/admin/docker-compose/docker-compose.yml up -d
	go run ./app/admin/cmd/migration
	$(MAKE) admin-web-build
	nunu run ./app/admin/cmd/server


admin-run:
	$(MAKE) admin-web-build
	nunu run ./app/admin/cmd/server

admin-migrate:
	go run ./app/admin/cmd/migration

admin-web-build:
	HUSKY=0 pnpm --dir ./app/admin/web install --frozen-lockfile
	pnpm --dir ./app/admin/web build

admin-build:
	$(MAKE) admin-web-build
	mkdir -p ./bin
	go build -ldflags="-s -w" -o ./bin/admin-server ./app/admin/cmd/server

admin-build-task:
	mkdir -p ./bin
	go build -ldflags="-s -w" -o ./bin/admin-task ./app/admin/cmd/task

admin-build-migration:
	mkdir -p ./bin
	go build -ldflags="-s -w" -o ./bin/admin-migration ./app/admin/cmd/migration

admin-test:
	go test ./app/admin/...

admin-mock:
	mkdir -p app/admin/test/mocks/service app/admin/test/mocks/repository
	mockgen -source=app/admin/internal/service/user.go -destination app/admin/test/mocks/service/user.go
	mockgen -source=app/admin/internal/repository/user.go -destination app/admin/test/mocks/repository/user.go
	mockgen -source=app/admin/internal/repository/repository.go -destination app/admin/test/mocks/repository/repository.go

admin-docker-task:
	docker build -f deploy/admin/build/Dockerfile --build-arg APP_RELATIVE_PATH=./app/admin/cmd/task --build-arg BUILD_ADMIN_WEB=false -t 1.1.1.1:5000/demo-task:v1 .
	docker run --rm -i 1.1.1.1:5000/demo-task:v1

admin-swag:
	swag init -g main.go -d ./app/admin/cmd/server,./app/admin/internal/handler,./app/admin/api/v1 --parseInternal -o ./app/admin/docs

.PHONY: home-run home-build home-test run-home build-home
home-run:
	go run ./app/home/cmd/server

home-build:
	mkdir -p ./bin
	go build -ldflags="-s -w" -o ./bin/home-server ./app/home/cmd/server

home-test:
	go test ./app/home/...

run-home: home-run

build-home: home-build

.PHONY: bootstrap build test mock docker swag build-all test-all verify
bootstrap: admin-bootstrap

build: admin-build

test: admin-test

mock: admin-mock

docker: admin-docker-task

swag: admin-swag

build-all: admin-build admin-build-task admin-build-migration home-build

test-all:
	go test ./...

verify: build-all test-all
