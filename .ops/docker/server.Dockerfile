FROM golang:1.21.13-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd cmd
COPY internal internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/datacat ./cmd/datacat

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/datacat /datacat
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/datacat"]
