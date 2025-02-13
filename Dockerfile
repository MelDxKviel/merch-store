FROM golang:1.23.6 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o merch-store ./cmd/server/main.go

FROM scratch
COPY --from=builder /app/merch-store /merch-store
EXPOSE 8080
ENTRYPOINT ["/merch-store"]
