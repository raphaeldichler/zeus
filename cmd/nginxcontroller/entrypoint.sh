#!/usr/bin/env sh

/docker-entrypoint.sh nginx -c /etc/zeus/ingress/nginx.conf -g 'daemon off;' &

while [ ! -f /var/run/nginx.pid ]; do sleep 0.1; done
NGINX_PID=$(cat /var/run/nginx.pid)
NGINX_PATH=$(which nginx)

echo "$NGINX_PID"
echo "$NGINX_PATH"

./opt/zeus/nginxcontroller/nginxcontroller --nginx

wait "$NGINX_PID":q
