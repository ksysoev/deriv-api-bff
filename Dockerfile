FROM golang:1.23 AS builder

WORKDIR /app/

COPY . .

RUN go mod download
RUN go mod verify

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /main ./cmd/bff

FROM scratch

COPY --from=builder /main /main

COPY --from=builder /app/runtime/config.yaml /runtime/config.yaml

ENTRYPOINT [ "./main", "server" ]
