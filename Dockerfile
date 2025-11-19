# Build React frontend
FROM node:20-alpine AS build_react_stage

WORKDIR /app/client

COPY client/package*.json ./
RUN npm ci --only=production && npm cache clean --force

COPY client/ ./
ARG REACT_APP_BACKEND_URL
ENV REACT_APP_BACKEND_URL=${REACT_APP_BACKEND_URL}
RUN npm run build

# Build Go backend
FROM golang:1.24-alpine AS build_go_stage

RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

WORKDIR /app/server

COPY server/go.mod server/go.sum ./
RUN go mod download && go mod verify

COPY server/ ./
RUN go build -ldflags="-w -s" -o seek-tune

# Final runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    ffmpeg \
    python3 \
    py3-pip \
    && pip3 install --no-cache-dir yt-dlp --break-system-packages

WORKDIR /app

COPY --from=build_go_stage /app/server/seek-tune .

RUN mkdir -p static
COPY --from=build_react_stage /app/client/build ./static

RUN mkdir -p db songs recordings snippets tmp && \
    chmod -R 755 db songs recordings snippets tmp

ENV ENV=production

EXPOSE 5000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:5000/ || exit 1

# Run as non-root user for security
RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser && \
    chown -R appuser:appuser /app

USER appuser

CMD ["./seek-tune", "serve", "http", "5000"]