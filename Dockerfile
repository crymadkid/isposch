# --- билд ---
FROM golang:1.27.1-alpine AS build
WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/isposch .

# --- итоговый образ ---
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/isposch /isposch

EXPOSE 8080
ENTRYPOINT ["/isposch"]
CMD ["-addr", ":8080"]