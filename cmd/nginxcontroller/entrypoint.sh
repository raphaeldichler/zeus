#!/usr/bin/env sh

./opt/zeus/nginxcontroller &

exec /docker-entrypoint.sh nginx -g 'daemon off;'
