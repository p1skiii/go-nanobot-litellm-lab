FROM golang:1.26 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/nanobot ./cmd/server

FROM scratch

WORKDIR /app
COPY --from=build /out/nanobot /app/nanobot
COPY configs /app/configs

EXPOSE 8080
ENTRYPOINT ["/app/nanobot"]
