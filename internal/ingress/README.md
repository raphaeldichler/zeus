
## Structure of internal nginx config

Depending on the internal configuration the config files of the nginx server is structured as followes.

```text
├ /etc/nginx/
├── nginx.conf 
├── sites/
│   ├── example.com#80.conf
|   ├──	example.com#443.conf
│   └── app.example.com#443.conf
├── locations/
│   ├── example.com#80/
│   │   ├── bG9jYXRpb24gLy53ZWxsLWtub3duL2FjbWUtY2hhbGxlbmdlLwo.conf
```


take care of all necessary interaction between the zeus-daemon and the nginx server.
To dont rely on docker exec we will use a lightweigh server which runs inside the nginx container which will
```text
├ /var/run/zeus/poseidon/ingress/connection.sock
```

zeus/nginx

easier to maintain the initial state of the nginx server.
we manage 
