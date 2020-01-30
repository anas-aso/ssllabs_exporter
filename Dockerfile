FROM golang:1.13.5-alpine3.10 AS builder

RUN apk update && \
    apk upgrade && \
    apk add --no-cache git

WORKDIR /workdir

# Download the dependecies first for faster iterations
COPY go.mod go.sum /workdir/
RUN go mod download

COPY . /workdir/

# Set the version to the tag, otherwise use the commit hash
RUN git describe --exact-match --tags HEAD > version || git rev-parse HEAD > version && cat version

RUN export VERSION=$(cat version) && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s -X main.version=${VERSION}" -o /workdir/ssllabs_exporter

# Create a "nobody" user for the next image
RUN echo "nobody:x:65534:65534:Nobody:/:" > /etc_passwd



FROM scratch

COPY --from=builder /workdir/ssllabs_exporter /bin/ssllabs_exporter
# Required for HTTPS requests done by the application
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Required to be able to run as a non-root user (nobody)
COPY --from=builder /etc_passwd /etc/passwd

USER nobody

ENTRYPOINT ["/bin/ssllabs_exporter"]
