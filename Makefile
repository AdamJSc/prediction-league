app.start:
	docker-compose up -d db assets

app.run:
	docker-compose up -d app && docker-compose logs -f

app.stop:
	docker-compose stop

app.kill:
	docker-compose down

test.start:
	docker-compose up -d db_test

test.run:
	docker-compose up app_test

test.stop:
	docker-compose stop app_test db_test
