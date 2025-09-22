FROM golang:1.21.13-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY api api
COPY cmd cmd
COPY internal internal
COPY pkg pkg
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/datacat ./cmd/datacat

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/datacat /datacat
EXPOSE 8080 4317
USER nonroot:nonroot
ENTRYPOINT ["/datacat"]
