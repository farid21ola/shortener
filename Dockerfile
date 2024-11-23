FROM golang:1.23.1

ENV GOPATH=/

COPY ./ ./

RUN apt-get update && apt-get install -y wget && \
    wget -O /tmp/dockerize.tar.gz https://github.com/jwilder/dockerize/releases/download/v0.6.1/dockerize-linux-amd64-v0.6.1.tar.gz && \
    tar -C /usr/local/bin -xzvf /tmp/dockerize.tar.gz && \
    chmod +x /usr/local/bin/dockerize && \
    rm /tmp/dockerize.tar.gz

RUN go mod download

RUN go build -o url-shortener ./cmd/shortener/main.go

CMD ["dockerize", "-wait", "tcp://db:5432", "-timeout", "60s", "./url-shortener"]