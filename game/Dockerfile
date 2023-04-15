ARG GOLANG_IMAGE_VERSION=1.20-alpine
FROM golang:${GOLANG_IMAGE_VERSION} AS builder
RUN pkg add --no-cache nodejs yarn
WORKDIR /go/src/x
ADD . ./
RUN make site

FROM alpine:3.17
COPY --from=builder /go/src/x/site /usr/local/bin/site
ENV HTTP_DEMO=simple-ui
CMD ["/usr/local/bin/site"]
