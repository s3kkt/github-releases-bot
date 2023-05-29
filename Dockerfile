FROM golang:1.20 as build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build

FROM debian:stable-slim

RUN apt-get update && \
    apt-get -y install ca-certificates && \
    apt-get clean && \
    rm -rf /var/cache/apt/
COPY --from=build /app/github-releases-bot .
COPY internal .

ENTRYPOINT ["/github-releases-bot"]