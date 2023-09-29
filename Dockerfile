FROM golang:1.21-alpine as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG version
RUN CGO_ENABLED=0 go build -a -ldflags "-w -s -X main.version=$version" -o valetudopng ./cmd/valetudopng/main.go

FROM scratch
COPY --from=builder /etc/ssl/cert.pem /etc/ssl/
COPY --from=builder /app/valetudopng /valetudopng
CMD ["/valetudopng"]
