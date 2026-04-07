up:
	docker-compose up --build

down:
	docker-compose down

restart:
	docker-compose down
	docker-compose up --build

build-graphql:
	docker-compose build graphql-gateway

logs:
	docker-compose logs -f

logs-graphql:
	docker-compose logs -f graphql-gateway
