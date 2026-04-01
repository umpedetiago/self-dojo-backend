FROM golang:1.25.5-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -trimpath -ldflags="-s -w" -o /out/go_api ./cmd

FROM alpine:3.21

RUN apk add --no-cache ca-certificates && update-ca-certificates

RUN addgroup -S app && adduser -S -G app app
USER app

WORKDIR /app
COPY --from=build /out/go_api /app/go_api
COPY --from=build /src/docs /app/docs

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/go_api"]

