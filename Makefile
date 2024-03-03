
postgres:
	docker run --name postgres -p 3005:5432 --rm -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

pgadmin:
	docker run --rm --name pgadmin -e "PGADMIN_DEFAULT_EMAIL=name@example.com" -e "PGADMIN_DEFAULT_PASSWORD=admin" -p 5050:80 -d dpage/pgadmin4 
	docker network create --driver bridge pgnetwork
	docker network connect pgnetwork pgadmin
	docker network connect pgnetwork postgres

rm_container:
	docker stop postgres
	docker stop pgadmin
	docker network rm pgnetwork
createdb:
	docker exec -it postgres createdb --username=root --owner=root simple_bank

migrate_up:
	migrate -path db/migration/ -database "postgresql://root:secret@localhost:3005/simple_bank?sslmode=disable" -verbose up

migrate_down:
	migrate -path db/migration/ -database "postgresql://root:secret@localhost:3005/simple_bank?sslmode=disable" -verbose down

dropdb:
	docker exec -it postgres dropdb simple_bank

sqlc:
	sqlc generate


test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -destination db/mock/store.go -package mockdb  simplebank/db/sqlc Store

.PHONY: createdb dropdb postgres pgadmin rm_container migrate_up migrate_down test server mock
