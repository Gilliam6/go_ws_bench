FROM golang:1.23.0-alpine

RUN apk update && apk add --no-cache \
    build-base \
    curl

WORKDIR /app
COPY . .
RUN go mod download
RUN #go build -o ./stress /app/cmd/gorilla/main.go
RUN #go build -o ./stress /app/cmd/fasthttp/main.go
RUN #go build -o ./stress /app/cmd/std/main.go
RUN go build -o ./stress /app/cmd/broadcaster/main.go
ENV GOMEMLIMIT=100MiB
EXPOSE 4242
CMD ./stress