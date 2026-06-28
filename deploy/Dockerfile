# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS go-builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/aitsd ./cmd/server \
 && CGO_ENABLED=0 GOOS=linux go build -o /out/aitsd-migrate ./cmd/migrate \
 && CGO_ENABLED=0 GOOS=linux go build -o /out/aitsd-worker ./cmd/worker

FROM node:22-alpine AS web-builder
WORKDIR /src/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM alpine:3.20 AS runtime
WORKDIR /srv
RUN apk add --no-cache ca-certificates tzdata
ENV PATH="/srv/bin:${PATH}"
COPY --from=go-builder /out/aitsd /srv/bin/aitsd
COPY --from=go-builder /out/aitsd-migrate /srv/bin/aitsd-migrate
COPY --from=go-builder /out/aitsd-worker /srv/bin/aitsd-worker
COPY --from=web-builder /src/web/dist /srv/web/dist
COPY deploy/entrypoint.sh /srv/bin/aitsd-entrypoint
RUN chmod +x /srv/bin/aitsd-entrypoint && mkdir -p /srv/data
EXPOSE 40680
ENTRYPOINT ["aitsd-entrypoint"]
