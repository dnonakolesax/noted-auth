# Установка модулей и тесты
FROM golang:1.24.7 as modules

ADD go.mod go.sum /m/
RUN cd /m && go mod download

RUN make test

# Сборка приложения
FROM golang:1.24.7 as builder

COPY --from=modules /go/pkg /go/pkg

# Пользователь без прав
RUN useradd -u 10001 auth-runner

RUN mkdir -p /noted-auth
ADD . /noted-auth
WORKDIR /noted-auth

# Сборка
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -o ./bin/noted-auth ./cmd/api

# Запуск в пустом контейнере
FROM scratch

# Копируем пользователя без прав с прошлого этапа
COPY --from=builder /etc/passwd /etc/passwd
# Запускаем от имени этого пользователя
USER auth-runner

COPY --from=builder /noted-auth/bin/noted-auth /noted-auth

CMD ["/noted-auth"]