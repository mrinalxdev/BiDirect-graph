FROM golang:1.21-alpine

WORKDIR /app

RUN apk add --no-cache gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o graph-server cmd/server/main.go
CMD ["./graph-server"]
