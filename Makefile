start:
	docker-compose up -d app && docker-compose logs -f

restart:
	docker-compose down && make start
