FROM golang:1.26.2-alpine3.22

RUN addgroup -S devgroup && adduser -S devuser -G devgroup -s /bin/ash

WORKDIR /workspace

RUN chown -R devuser:devgroup /workspace

USER devuser

CMD ["sleep", "infinity"]
