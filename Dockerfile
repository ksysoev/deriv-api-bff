FROM golang:1.23 AS builder

ARG VERSION=${VERSION}
ARG BUILD=${BUILD}

WORKDIR /app/

COPY . .

RUN go mod download
RUN go mod verify

RUN CGO_ENABLED=0 go build -o /deriv-api-bff  -ldflags "-X main.version=$VERSION -X main.build=$BUILD" ./cmd/bff

FROM scratch

COPY --from=builder /deriv-api-bff /deriv-api-bff

COPY --from=builder /app/runtime/config.yaml /runtime/config.yaml

ENTRYPOINT [ "./deriv-api-bff" ]
CMD [ "server" ]
