FROM golang:1.23.0 AS builder

WORKDIR /src/golox

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o ./build/glox ./cmd/...

FROM gcr.io/distroless/base:latest

COPY --from=builder /src/golox/build/glox /usr/local/bin/glox

CMD ["glox"]
