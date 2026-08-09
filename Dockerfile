FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/url-shortener .

FROM alpine:3.24
COPY --from=build /out/url-shortener /url-shortener
USER 65534:65534
EXPOSE 8080
ENTRYPOINT ["/url-shortener"]