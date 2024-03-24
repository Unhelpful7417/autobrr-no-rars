FROM golang:alpine3.19 AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

# I would have liked to statically link this binary and run in a scratch base image, but
# I keep getting "runtime error: invalid memory address or nil pointer dereference". So
# I'll leave this here in case I figure it out for some reason:
# RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /autobrr-no-rars .
RUN CGO_ENABLED=0 go build -o /autobrr-no-rars .

FROM alpine:3.19

COPY --from=build --chmod=755 "/autobrr-no-rars" "/autobrr-no-rars"

ENTRYPOINT [ "/autobrr-no-rars" ]