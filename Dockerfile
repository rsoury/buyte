FROM golang:1.12.7-alpine

ARG BUILD_ENV="development"

# Install git and essentials.
# Git is required for fetching the dependencies.
# sdk/essentials is required for make
RUN apk update && \
    apk add --no-cache git && \
    apk add --no-cache alpine-sdk && \
    apk add --no-cache openssl-dev && \
    rm -rf /var/cache/apk/*

# Create a group and user
# RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Get Dep and CFSSL
RUN go get -u github.com/golang/dep/cmd/dep
# We only need cfssl to generate Apple Pay certs
# RUN go get -u github.com/cloudflare/cfssl/cmd/...

WORKDIR $GOPATH/src/buyte/
COPY . .
# COPY ./certs /certs
# COPY ./config.${BUILD_ENV}.yaml ./config.yaml

# Set ENV variables for container
ENV APP_ENV = ${BUILD_ENV}
ENV AWS_ACCESS_KEY_ID=""
ENV AWS_SECRET_ACCESS_KEY=""
ENV SERVER_CONTAINER=true
ENV PKG_CONFIG_PATH=/usr/lib/pkgconfig
ENV PORT=80
# ENV APPLE_CERTS="/"

# Make Container Build -- Fetches dependencies and Builds Binary for Alpine.
RUN make container-build

RUN chmod -R +r ./certs
# Use an unprivileged user.
# USER appuser

EXPOSE 80

# Run the API
CMD ["./bin/buyte", "api", "-v"]