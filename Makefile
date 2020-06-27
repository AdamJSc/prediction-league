app.run:
	docker-compose up -d db \
	&& npm install \
	&& npm run watch \
	| go run service/main.go

app.up:
	docker-compose up -d app && docker-compose logs -f

app.stop:
	docker-compose stop

app.restart:
	docker-compose stop app && docker-compose up -d app

app.kill:
	docker-compose down

app.build:
	# go mod vendor
	docker run --rm -v ${PWD}:/app -w /app golang:1.14-alpine go mod vendor

	# package static assets
	docker build --tag prediction-league-pkger:v1 ./docker/images/pkger
	docker run --rm -v ${PWD}:/app prediction-league-pkger:v1 pkger -include /service/views/html
	mv pkged.go service/pkged.go

	# go build
	docker run --rm -v ${PWD}:/app -w /app -e CGO_ENABLED="0" golang:1.14-alpine go build -mod vendor -o ./bin/prediction-league ./service

	# clean up
	rm -rf vendor

test.run:
	docker-compose up -d db_test && go test -v ./...

test.up:
	docker-compose up app_test

test.stop:
	docker-compose stop app_test db_test
