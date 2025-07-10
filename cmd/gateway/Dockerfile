FROM golang:1.21-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o gateway ./cmd/gateway

FROM alpine
COPY --from=build /app/gateway /service
ENTRYPOINT ["/service"]
