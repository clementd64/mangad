FROM golang:1.17-alpine as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o mangad ./cmd/mangad

FROM gcr.io/distroless/static

LABEL org.opencontainers.image.source https://github.com/clementd64/mangad

COPY --from=builder /app/mangad /mangad
ENTRYPOINT [ "/mangad" ]