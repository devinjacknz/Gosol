FROM golang:1.21-alpine

WORKDIR /app

RUN apk add --no-cache gcc musl-dev curl wget

COPY go.mod go.sum ./
RUN go get github.com/gorilla/websocket && go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o main .

EXPOSE 8080

# Add healthcheck dependencies and ensure proper permissions
RUN chmod +x main

CMD ["./main"]
