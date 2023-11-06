FROM golang:1.21.2 AS builder

WORKDIR /src/golox

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o ./build/glox ./cmd/...

FROM dart:2.19.4 AS dart

RUN apt-get update && apt-get install -y git
RUN git clone https://github.com/munificent/craftinginterpreters.git

WORKDIR craftinginterpreters

RUN cd tool && dart pub get
COPY test.sh .

COPY --from=builder /src/golox/build/glox /usr/local/bin/glox

CMD ["./test.sh"]

