############################
# STEP 1 build executable binary
############################
FROM golang:alpine as builder
# FROM golang@sha256:8dea7186cf96e6072c23bcbac842d140fe0186758bcc215acb1745f584984857 as builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && \
    apk upgrade && \ 
    apk add --no-cache git ca-certificates && \
    rm -rf /var/cache/apk/* && \
    update-ca-certificates


# Create appuser
RUN adduser -D -g '' appuser

WORKDIR /revese-proxy/
COPY . .

# Fetch dependencies.
RUN go mod download
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/reverse-proxy

############################
# STEP 2 build a small image
############################
FROM scratch

# Import from builder.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /revese-proxy/bin/reverse-proxy ./

# Use an unprivileged user.
USER appuser

EXPOSE 9000

ENTRYPOINT ["./reverse-proxy"]