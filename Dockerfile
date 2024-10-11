FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /app/cmd/bff/cmd ./cmd/bff

FROM scratch

COPY --from=builder /app/cmd/bff/cmd /cmd

CMD ["/cmd"]
