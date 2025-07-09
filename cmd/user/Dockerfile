FROM golang:1.21-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o user ./cmd/user

FROM alpine
COPY --from=build /app/user /service
ENTRYPOINT ["/service"]
