FROM golang:latest AS builder
WORKDIR /app
COPY go.mod .
COPY main.go .
COPY pipeline/ ./pipeline/
RUN go build -o pipeline .

FROM alpine:latest
LABEL version="v1.0.0"
WORKDIR /root/
COPY --from=builder /app/pipeline .
ENTRYPOINT ["./pipeline"]