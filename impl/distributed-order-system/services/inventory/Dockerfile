FROM golang:1.26.2-alpine3.22 AS builder

RUN addgroup -S appgroup \
    && adduser -S appuser -G appgroup

WORKDIR /workspace

COPY ./shared /workspace/shared/

COPY ./services/inventory/go.* /workspace/inventory/

WORKDIR /workspace/inventory/

RUN go mod download

COPY ./services/inventory/ /workspace/inventory/

RUN chown -R appuser:appgroup /workspace

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w"  -o ./build/main .

ENV PORT=8001

EXPOSE 8001

USER appuser

CMD ["/workspace/build/main"]
