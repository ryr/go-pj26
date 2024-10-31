FROM golang AS builder
RUN mkdir -p /go/src/pipeline
WORKDIR /go/src/pipeline
ADD main.go .
ADD go.mod .
RUN go install .

FROM alpine:latest
LABEL version="1.0.0"
LABEL maintainer="ryr<test@test.ru>"
WORKDIR /root/
COPY --from=builder /go/bin/pipeline .
ENTRYPOINT ./pipeline
EXPOSE 8080
