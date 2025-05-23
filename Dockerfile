FROM golang:1.23.2 AS builder
WORKDIR /go/src
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

COPY . .
RUN go mod tidy
RUN go build -o /go/bin/app ./cmd/main.go

FROM alpine:latest
RUN apk add --no-cache ffmpeg freetype freetype-dev fontconfig ttf-dejavu
WORKDIR /app
COPY --from=builder /go/bin/app /app/app
COPY .env /app/.env
RUN chmod +x /app/app
RUN ls -lh /app
CMD ["/app/app"]
