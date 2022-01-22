# Dependencies
FROM golang:1.13-alpine AS deps

ENV APP_HOME=/app
ENV GOOS=linux
ENV GOARCH=amd64

RUN mkdir -p $APP_HOME
WORKDIR $APP_HOME

COPY go.mod .
COPY go.sum .

RUN apk --no-cache add --virtual .deps git && \
    go mod download && \
    apk del .deps

# Builder
FROM deps AS builder

COPY . .

RUN apk --no-cache add --virtual .deps g++ make && \
    make && \
    apk del .deps

# Runner
FROM golang:1.13-alpine

ENV APP_HOME=/app
ENV CONSUL_TMPL_VERSION=0.22.0
ENV MIGRATE_VERSION=4.8.0
RUN apk --no-cache add curl && \
    curl -sL "https://releases.hashicorp.com/consul-template/${CONSUL_TMPL_VERSION}/consul-template_${CONSUL_TMPL_VERSION}_linux_amd64.tgz" | tar -C /usr/bin -xvz && \
    chmod a+x /usr/bin/consul-template && \
    curl -L "https://github.com/golang-migrate/migrate/releases/download/v${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz" | tar xz -C /tmp && \
    cp /tmp/migrate.linux-amd64 /usr/local/bin/migrate




WORKDIR $APP_HOME

COPY --from=builder $APP_HOME/entrypoint.sh ./entrypoint.sh
COPY --from=builder $APP_HOME/consul ./consul
COPY --from=builder $APP_HOME/dist/* ./
COPY ./db/migrations/* ./
COPY ./docs/* ./docs/

# Expose port
EXPOSE 80

ENTRYPOINT ["sh", "./entrypoint.sh"]
