FROM golang:1.21-bullseye as builder

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" go build -o app

FROM alpine:latest
RUN mkdir /app
WORKDIR /app
COPY --from=builder /app/app .
EXPOSE 8080
EXPOSE 8081
ENTRYPOINT ["./app"]