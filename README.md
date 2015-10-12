## Comaha update server for coreos

DISCLAIMER: "COREOS" is a trademark of CoreOS, Inc.

### Nginx reverse proxy config
#### Endpoints
Basically, the following endpoints need to be proxied for proper operation:
 - `/file`
 - `/update`
 - `/panel`

#### Headers
Set the following headers for `/update` endpoint:
```
proxy_set_header X-Forwarded-Proto $scheme;
proxy_set_header Host $host:$server_port;
```
It will enable assembling proper URLs for local file storage.


Set this for all endpoints to enable proper remote address logging:
```
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
```
