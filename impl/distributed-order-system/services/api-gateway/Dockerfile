FROM golang:1.26.2-alpine3.22 AS builder

RUN addgroup -S appgroup \
    && adduser -S appuser -G appgroup

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY ./ ./

RUN chown -R appuser:appgroup /app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w"  -o ./build/main .

ENV PORT=8080

EXPOSE 8080

USER appuser

CMD ["/app/build/main"]
