FROM golang:1.17-buster AS build
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o sia_exporter

FROM debian:buster-slim
COPY --from=build --chmod=755 /app/sia_exporter /sia_exporter
CMD ["/sia_exporter"]
