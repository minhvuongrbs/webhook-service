docker-compose-up:
	docker-compose up -d

generate_sqlc:## generate database model
	sqlc generate -f ./internal/adapters/repository/sqlc/sqlc.yaml

generate_proto:## generate proto
	buf generate

coverage:
	go test -coverprofile=coverage.out ./...
