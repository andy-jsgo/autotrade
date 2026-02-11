FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api

FROM alpine:3.20
RUN adduser -D app
USER app
COPY --from=builder /bin/api /bin/api
EXPOSE 8080 8081
ENTRYPOINT ["/bin/api"]
