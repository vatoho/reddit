FROM golang:1.21.1

WORKDIR /usr/local/src

COPY ./ ./

RUN go mod tidy
RUN go build -o ./app_start ./cmd/redditclone/main.go

CMD ["./app_start"]
