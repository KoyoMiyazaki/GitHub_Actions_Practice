postgres:
	docker run --name postgres13 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -p 5555:5432 -d postgres:13-alpine

createdb:
	docker exec -it postgres13 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres13 dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5555/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5555/simple_bank?sslmode=disable" -verbose down

test:
	go test -v -cover ./...

.PHONY: postgres createdb dropdb migrateup migratedown