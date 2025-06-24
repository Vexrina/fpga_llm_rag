up:
	docker-compose -f float-weaver/docker-compose.yml -f rag/docker-compose.yml up --build

down:
	docker-compose -f float-weaver/docker-compose.yml -f rag/docker-compose.yml down

restart:
	docker-compose -f float-weaver/docker-compose.yml -f rag/docker-compose.yml down
	docker-compose -f float-weaver/docker-compose.yml -f rag/docker-compose.yml up --build 