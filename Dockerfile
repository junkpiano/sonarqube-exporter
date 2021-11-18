# syntax=docker/dockerfile:1

FROM golang:1.16

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN go build -o sonarqube-exporter cmd/main.go

CMD ["./sonarqube-exporter"]