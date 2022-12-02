FROM golang:1.18-alpine AS base
RUN apk update && apk upgrade && apk add gcc g++
RUN mkdir /workdir && mkdir /build

WORKDIR /workdir
COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN go build -o /build/app ./cmd/build/

FROM base AS test
RUN addgroup -S test && adduser -S test -G test
USER test
CMD ["scripts/verify.sh"]

FROM alpine:3.14 AS app
RUN apk update && apk upgrade && addgroup -S app && adduser -S app -G app
USER app
WORKDIR /workdir
COPY --from=base /build/app ./app
ENTRYPOINT ["./app"]
