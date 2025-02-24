postgres:
	docker run --name auth-service-bd -p 5433:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:17-alpine

createdb:
	docker exec -it auth-service-bd createdb --username=root --owner=root auth 

dropdb:
	docker exec -it auth-service-bd dropdb --username=root auth 

migrateup:
	migrate -path internal/db/migration -database "postgresql://root:secret@localhost:5433/auth?sslmode=disable" -verbose up

migratedown:
	migrate -path internal/db/migration -database "postgresql://root:secret@localhost:5433/auth?sslmode=disable" -verbose down 

.PHONY: postgres createdb dropdb migrateup migratedown 