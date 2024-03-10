

DB_URL=postgresql://root:secret@localhost:3005/simple_bank?sslmode=disable


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
	migrate -path db/migration/ -database "$(DB_URL)" -verbose up

migrate_up1:
	migrate -path db/migration/ -database "$(DB_URL)" -verbose up 1

migrate_down:
	migrate -path db/migration/ -database "$(DB_URL)" -verbose down

migrate_down1:
	migrate -path db/migration/ -database "$(DB_URL)" -verbose down 1

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

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)


proto:
	rm -f pb/*.go
	rm -f doc/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb  --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=simple_bank \
	proto/*.proto
	statik -src=./doc/swagger -dest=./doc
evans:
	evans --host localhost --port 9090 -r repl

.PHONY: createdb dropdb postgres pgadmin rm_container migrate_up migrate_down migrate_up1 migrate_down1 test server mock proto evans
