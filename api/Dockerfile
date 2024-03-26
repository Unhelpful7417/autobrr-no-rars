FROM golang:alpine3.19 AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN go build -o /autobrr-no-rars .

# This base image contains CA certificates we need for outbound HTTPS requests
# https://github.com/GoogleContainerTools/distroless/blob/main/base/README.md
FROM gcr.io/distroless/static

COPY --from=build --chmod=755 "/autobrr-no-rars" "/autobrr-no-rars"

ENTRYPOINT [ "/autobrr-no-rars" ]
