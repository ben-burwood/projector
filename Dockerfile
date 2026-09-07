FROM node:latest AS build-vue

WORKDIR /app/frontend

COPY frontend/package*.json ./

RUN npm install

COPY frontend/. .

RUN npm run build-only

FROM --platform=$BUILDPLATFORM golang:alpine AS build-go

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev

# ca-certificates provides /etc/ssl/certs/ca-certificates.crt copied into scratch.
RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# The built SPA must be present before `go build` so //go:embed frontend/dist works.
COPY --from=build-vue /app/frontend/dist ./frontend/dist

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags "-s -w -X main.version=${VERSION}" -o /server

FROM scratch

WORKDIR /

COPY --from=build-go /server /server

# scratch has no CA bundle; outbound HTTPS to provider APIs fails without this.
COPY --from=build-go /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE 8080

COPY --from=ghcr.io/tarampampam/microcheck:1 /bin/httpcheck /bin/httpcheck
HEALTHCHECK --interval=30s --timeout=3s --retries=3 --start-period=10s CMD ["httpcheck", "http://localhost:8080"]

ENTRYPOINT ["/server"]
