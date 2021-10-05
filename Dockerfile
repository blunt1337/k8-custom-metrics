#
# Go builder
#
FROM golang:1.17.1-alpine3.14 as builder

COPY . /app
RUN cd /app && CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/metrics-apiserver

#
# Final container
#
FROM scratch

COPY --from=builder /app/metrics-apiserver /app/metrics-apiserver

# Start
WORKDIR /app
CMD ["/app/metrics-apiserver", "--logtostderr"]
ENTRYPOINT ["/app/metrics-apiserver", "--logtostderr"]
