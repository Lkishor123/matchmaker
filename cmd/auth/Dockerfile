FROM golang:1.21-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o auth ./cmd/auth

FROM alpine
COPY --from=build /app/auth /service
ENTRYPOINT ["/service"]
