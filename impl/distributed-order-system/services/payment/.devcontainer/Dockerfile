FROM golang:1.26.2-alpine3.22

RUN apk add --no-cache git \
    && go install github.com/air-verse/air@latest

CMD ["sleep", "infinity"]
