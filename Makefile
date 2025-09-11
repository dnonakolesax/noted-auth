.PHONY: run

run:
	go run cmd/api/main.go

swagger-docs:
	swag init -g main.go -d cmd/api,internal/delivery/auth/v1,internal/model,internal/delivery/user/v1  

swagger-server:
	go run cmd/swagger/main.go
