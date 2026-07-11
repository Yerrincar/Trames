ARG NODE_IMAGE=node:22-alpine3.22
ARG GO_IMAGE=golang:1.25-alpine3.22
ARG RUNTIME_IMAGE=alpine:3.22

# Vue front

FROM ${NODE_IMAGE} AS front-build

WORKDIR /src/web

COPY web/package*.json ./
RUN npm ci

COPY web/ ./      
RUN npm run build


# App and goose

FROM ${GO_IMAGE} AS go-build

ARG GOOSE_VERSION=v3.27.2

RUN apk add --no-cache  build-base

WORKDIR /src

COPY go.mod go.sum /   
RUN go mod download


COPY . .

COPY --from=front-build /src/web/dist ./web/dist

RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/trames ./cmd

RUN CGO_ENABLED=1 GOBIN=/out go install github.com/pressly/goose/v3/cmd/goose@${GOOSE_VERSION}

# Runtime image

FROM ${RUNTIME_IMAGE}

RUN apk add --no-cache ca-certificates libgcc sqlite tzdata && addgroup -S -g 10001 trames && adduser -S -D -H -u 10001 -G trames trames

WORKDIR /app 

COPY --from=go-build --chown=10001:10001 /out/trames /app/trames
COPY --from=go-build --chown=10001:10001 /out/goose /app/goose
COPY --from=front-build --chown=10001:10001 /src/web/dist /app/web/dist

COPY --chown=10001:10001 internal/core/platform/storage/migrations /app/migrations

RUN mkdir -p /data /config /tmp \
    && chown 10001:10001 /data /config /tmp \
    && ln -s /data/trames.db /app/trames.db \
    && ln -s /config/.env /app/.env

USER 10001:10001

EXPOSE 4040

ENTRYPOINT [ "/app/trames" ]





