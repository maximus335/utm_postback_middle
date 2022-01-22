#!/bin/sh

/usr/bin/consul-template \
  -consul-addr "${CONSUL_ADDR}" \
  -consul-auth "${CONSUL_AUTH}" \
  -consul-token "${CONSUL_TOKEN}" \
  -config "${APP_HOME}/consul/config.hcl" \
  -exec "$@"
