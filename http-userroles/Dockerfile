FROM golang:1.20-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server ./server
RUN go build -o cli ./cli


FROM alpine:3.14 AS server

COPY --from=builder /build/server/server /bin/server
ENTRYPOINT ["/bin/server"]


FROM alpine:3.14 AS cli

COPY --from=builder /build/cli/cli /bin/cli
CMD ["/bin/cli"]
