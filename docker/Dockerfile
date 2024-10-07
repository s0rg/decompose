FROM golang:1.23 AS builder

ADD .. /go/src/github.com/s0rg/decompose
WORKDIR /go/src/github.com/s0rg/decompose

RUN make build

FROM scratch

ARG BUILD_DATE
ARG BUILD_REV

COPY --from=builder /go/src/github.com/s0rg/decompose/bin/decompose /decompose

ENTRYPOINT ["/decompose"]

LABEL org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.revision="${BUILD_REV}" \
      org.opencontainers.image.title="decompose" \
      org.opencontainers.image.authors="s0rg" \
      org.opencontainers.image.vendor="s0rg" \
      org.opencontainers.image.source="https://github.com/s0rg/decompose"
