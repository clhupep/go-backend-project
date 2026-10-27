FROM golang:1.26 AS dependencies
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
FROM dependencies AS build
COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api
FROM alpine:3.22
RUN adduser -D -u 10001 app
USER app
COPY --from=build /api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
