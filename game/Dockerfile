ARG GOLANG_IMAGE_VERSION=1.20.3-alpine3.17
FROM golang:${GOLANG_IMAGE_VERSION} AS builder
RUN apk add --no-cache git make nodejs yarn
WORKDIR /go/src/x
ADD . ./
RUN make site

FROM alpine:3.17
COPY --from=builder /go/src/x/site /usr/local/bin/site
ENV HTTP_ADDR=0.0.0.0:3000
ENV HTTP_DEMO=simple-ui
EXPOSE 3000
CMD ["/usr/local/bin/site"]
