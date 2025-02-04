FROM golang:1.23 AS builder

WORKDIR /build
COPY . /build
RUN \
  --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 go build -ldflags="-w -s" -o vocabmemo ./cmd/vocabmemo

FROM alpine
COPY --from=builder /build/vocabmemo /bin/vocabmemo

WORKDIR /opt/vocabmemo
ENTRYPOINT [ "/bin/vocabmemo" ]
