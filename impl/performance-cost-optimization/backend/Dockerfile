FROM golang:1.26.2-bookworm AS build

WORKDIR /workspace

COPY shared/ shared/

COPY services/web-backend/go.mod services/web-backend/go.sum services/web-backend/
WORKDIR /workspace/services/web-backend
RUN go mod download

COPY services/mobile-backend/go.mod services/mobile-backend/go.sum services/mobile-backend/
WORKDIR /workspace/services/mobile-backend
RUN go mod download

COPY services/workers/go.mod services/workers/
WORKDIR /workspace/services/workers
RUN go mod tidy && go mod download

COPY services/web-backend/ services/web-backend/
COPY services/mobile-backend/ services/mobile-backend/
COPY services/workers/ services/workers/

WORKDIR /workspace/services/web-backend
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /workspace/bin/web-backend .

WORKDIR /workspace/services/mobile-backend
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /workspace/bin/mobile-backend .

WORKDIR /workspace/services/workers
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /workspace/bin/workers .

FROM golang:1.26.2-alpine

RUN apk add --no-cache ca-certificates && \
    adduser -D -H -s /sbin/nologin appuser

COPY --from=build /workspace/bin/web-backend /usr/local/bin/
COPY --from=build /workspace/bin/mobile-backend /usr/local/bin/
COPY --from=build /workspace/bin/workers /usr/local/bin/
COPY entrypoint.sh /usr/local/bin/

RUN chmod +x /usr/local/bin/entrypoint.sh

USER appuser

EXPOSE 8080 8081

ENTRYPOINT ["entrypoint.sh"]
