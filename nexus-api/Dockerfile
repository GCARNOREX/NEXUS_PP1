FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/nexus-api .

FROM alpine:3.20
COPY --from=build /out/nexus-api /usr/local/bin/nexus-api
EXPOSE 8080
CMD ["nexus-api"]