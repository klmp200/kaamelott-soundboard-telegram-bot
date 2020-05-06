FROM golang:1.14 AS builder

RUN mkdir /build
WORKDIR /build

# Copy the code from the host and compile it
COPY . .

RUN go build

RUN mkdir res
COPY sounds res/sounds

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /app .

# We use Alpine for it's ca-certificates needed by http lib
FROM alpine:latest
RUN apk add --no-cache ca-certificates apache2-utils
COPY --from=builder /app ./
COPY --from=builder /build/res ./


ENTRYPOINT ["./app"]
