FROM golang:alpine AS builder
WORKDIR /app
ADD . ./
RUN go mod tidy
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s"  -o echo-server .
RUN ls -al

FROM alpine:3
COPY --from=builder /app/echo-server /app/echo-server
EXPOSE 8080
ENTRYPOINT ["/app/echo-server"]