FROM golang:1.14-alpine as builder

ENV USER=app
ENV UID=1000

RUN adduser \    
    --disabled-password \    
    --gecos "" \    
    --home "/nonexistent" \
    --shell "/sbin/nologin" \    
    --no-create-home \    
    --uid "${UID}" \    
    "${USER}"

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /go/src/app

# Download and cache dependencies
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Build
COPY . .
RUN go build -ldflags "-w -s -X main.BUILD_TIME=$(date -u '+%Y%m%d_%H%M%S')" -o ./server.bin ./server

# Create dist
RUN mkdir -p dist/cert \
    && cp server.bin dist/server.bin \
    && cp areyouin.example.yaml dist/areyouin.yaml

# Final image
# FROM alpine:3.11
# RUN apk add --no-cache bash
FROM scratch
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --chown=app:app --from=builder /go/src/app/dist /app
WORKDIR /app
USER app:app
ENTRYPOINT ["/app/server.bin"]
EXPOSE 1822 2022 40186