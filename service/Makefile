DB_URL = postgres://${NODE_UTIL_DB_USERNAME}:${NODE_UTIL_DB_PASSWORD}@${NODE_UTIL_DB_ENDPOINT}/node_util?sslmode=disable

mock:
	go install github.com/golang/mock/mockgen@latest
	mockgen -package=store -destination=./db/store_mock.go -source=./db/store.go store

sqlc:
	docker pull kjconroy/sqlc
	#docker run --rm -v ${PWD}:/src kjconroy/sqlc sqlc generate
	docker run --rm -v "${PWD}:/src" -w /src kjconroy/sqlc:1.17.0 /workspace/sqlc generate

migrate_down:
	echo "y" | docker run --rm -v /db/migration/:/migrations  --network host migrate/migrate -path=/migrations -verbose -database=$(DB_URL) down -all 

migrate_drop:
	echo "y" | docker run --rm -v /db/migration/:/migrations  --network host migrate/migrate -path=/migrations -verbose -database=$(DB_URL) drop -f 

migrate_up:
	echo "y" | docker run --rm -v /db/migration/:/migrations  --network host migrate/migrate -path=/migrations -verbose -database=$(DB_URL) up 

.PHONY: sqlc mock migrate_down migrate_drop migrate_up