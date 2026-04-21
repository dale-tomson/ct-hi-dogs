# Stage 1: Build the React frontend
FROM node:22-alpine AS frontend-builder
RUN corepack enable pnpm
WORKDIR /app/frontend
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm run build

# Stage 2: Build the Go binary (with frontend dist already included)
FROM golang:1.26-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
# First copy the frontend dist from host  
COPY frontend/dist ./frontend/dist
# Then copy all source code
COPY . .
RUN go build -o dogs-api .

# Stage 3: Minimal production image
FROM alpine:3.21
WORKDIR /app
COPY --from=go-builder /app/dogs-api .
COPY --from=go-builder /app/frontend/dist ./frontend/dist
COPY dogs.json .
EXPOSE 8080
CMD ["./dogs-api"]
