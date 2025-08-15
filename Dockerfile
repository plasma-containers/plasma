FROM golang:1.24.6-alpine3.22 AS builder

WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY . .
RUN go build -o plasma-server main.go && rm -rf /tmp/* && rm -rf /tmp/*.*

FROM scratch

COPY --from=builder /app/plasma-server /
COPY --from=builder /tmp /tmp
EXPOSE 8080
CMD ["/plasma-server"]
