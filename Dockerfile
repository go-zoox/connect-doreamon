# Builder
FROM whatwewant/builder-go:v1.20-1 as builder

RUN apk add gcc g++ make

WORKDIR /app

COPY go.mod ./

COPY go.sum ./

RUN go mod download

COPY . ./

# RUN         CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -v -o server cmd/main.go

# 'CGO_ENABLED=0', go-sqlite3 requires cgo to work.
# RUN         go build -ldflags="-w -s" -v -o server cmd/main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -v -o server main.go

# Server
# FROM  scratch # x509: certificate signed by unknown authority
FROM whatwewant/alpine:v1.0.0

LABEL MAINTAINER="Zero<tobewhatwewant@gmail.com>"

WORKDIR /app

ARG VERSION=v1

COPY --from=builder /app/server /bin

EXPOSE 8080

ENV VERSION=${VERSION}

COPY ./entrypoint.sh /entrypoint.sh

CMD /entrypoint.sh
