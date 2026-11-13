# Frontend
FROM node:22-alpine AS web
WORKDIR /web
COPY web/package.json ./
RUN npm install
COPY web/ ./
RUN npm run build

# Backend
FROM golang:1.22-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY pkg ./pkg
COPY cmd/server ./cmd/server
COPY --from=web /web/dist ./web/dist
RUN CGO_ENABLED=0 go build -o /fuzzy-server ./cmd/server

# Runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
RUN mkdir -p /app/data
COPY --from=backend /fuzzy-server /app/fuzzy-server
COPY --from=web /web/dist /app/web/dist
ENV ADDR=:8080
ENV STATIC_DIR=/app/web/dist
ENV LOG_PATH=/app/data/fuzzy_journal.log
EXPOSE 8080
VOLUME ["/app/data"]
CMD ["/app/fuzzy-server"]
