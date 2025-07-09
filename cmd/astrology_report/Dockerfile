FROM golang:1.21-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o astrology_report ./cmd/astrology_report

FROM alpine
COPY --from=build /app/astrology_report /service
ENTRYPOINT ["/service"]
