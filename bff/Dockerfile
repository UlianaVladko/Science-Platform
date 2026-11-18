FROM golang:1.22-alpine AS build
WORKDIR /src

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

ENV CGO_ENABLED=0 GOOS=linux
RUN go build -trimpath -ldflags="-s -w" -o /out/bff ./cmd/server

FROM alpine:3.20
RUN adduser -D -H -u 10001 app
COPY --from=build /out/bff /usr/local/bin/bff

USER app
ENV PORT=8080 \
    UPSTREAM_URL=http://api:8000 \
    SYNTHETIC_NLP_READY=false
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/bff"]
