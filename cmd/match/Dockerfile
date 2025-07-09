FROM golang:1.21-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o match ./cmd/match

FROM alpine
COPY --from=build /app/match /service
ENTRYPOINT ["/service"]
