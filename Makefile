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

test.run:
	docker-compose up -d db_test && go test -v ./...

test.up:
	docker-compose up app_test

test.stop:
	docker-compose stop app_test db_test
