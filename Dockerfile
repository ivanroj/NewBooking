# syntax=docker/dockerfile:1

FROM golang:1.25-bookworm AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build --chown=nonroot:nonroot /out/server /server

USER nonroot:nonroot
ENV PORT=:8080
EXPOSE 8080
ENTRYPOINT ["/server"]
