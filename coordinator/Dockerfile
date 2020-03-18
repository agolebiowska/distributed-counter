## First stage
FROM golang:alpine AS builder

ENV CGO_ENABLED 0

WORKDIR /go/src
COPY . .

RUN apk add --no-cache git
RUN go get ./...

RUN GOOS=linux go build -a -installsuffix cgo -o server main.go

## Second stage
FROM alpine

WORKDIR /
COPY --from=builder /go/src/ /
RUN chmod +x ./server

CMD ["./server"]