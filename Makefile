.PHONY: run

test:
	./scripts/sqlmap.sh

run:
	mkdir -p /var/log/noted-auth
	go run cmd/api/main.go

swagger-docs:
	swag init -g main.go -d cmd/api,internal/delivery/auth/v1,internal/model,internal/delivery/user/v1,internal/delivery/session/v1  

swagger-server:
	go run cmd/swagger/main.go

lint:
	golangci-lint -v run 

easyjson:
	easyjson -all internal/model/*.go
