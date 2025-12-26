#!/bin/sh
set -e
export BACKEND_URL=${BACKEND_URL:-http://networth-backend:8080}
# Ensure the conf.d directory exists and is writable
mkdir -p /etc/nginx/conf.d
chmod 755 /etc/nginx/conf.d
envsubst '${BACKEND_URL}' < /etc/nginx/templates/default.conf.template > /etc/nginx/conf.d/default.conf
exec nginx -g "daemon off;"

