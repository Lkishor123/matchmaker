FROM golang:1.21-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o chat ./cmd/chat

FROM alpine
COPY --from=build /app/chat /service
ENTRYPOINT ["/service"]
