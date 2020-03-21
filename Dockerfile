## First stage
FROM golang:alpine AS builder

ENV CGO_ENABLED 0

WORKDIR /src
COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o server

## Second stage
FROM alpine
RUN apk add curl
COPY --from=builder /src/server /src/
CMD ["/src/server"]