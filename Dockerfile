FROM golang:1.14-alpine as builder
ENV GO111MODULE=on
LABEL maintainer="Martin Koppehel <martin@mko.dev>"

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

FROM haproxy:2.1-alpine
COPY --from=builder /app/kingress /usr/local/sbin/kingress
ENTRYPOINT ["/usr/local/sbin/kingress"]
