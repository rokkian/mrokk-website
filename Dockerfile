# Build the Go binary in a separate stage so the runtime image stays small.
FROM golang:1.27-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gowebsite .

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=build /out/gowebsite ./gowebsite
COPY content.yaml ./content.yaml

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/gowebsite"]

