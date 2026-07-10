FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git build-base sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

FROM alpine:3.20

RUN apk add --no-cache sqlite-libs

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/db ./db

CMD ["./main"]
