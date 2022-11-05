# Media server

## Configuration

### Launcing flags
`media-server --mode=env` для старта в режиме ENV-переменных

`media-server --mode=yaml --config=example-config.yaml` для старта в режиме YAML-конфига

### YAML variables
Example of `.yaml` configuration is aviable into `example-config.yaml`.

### ENV variables
`MEDIA_SERVER_HOST` : set server host (example: `localhost`)

`MEDIA_SERVER_PORT` : set server port (example: `8082`)

`MEDIA_SERVER_PREFIX` : set specific prefix for all API handlers (example: `/media-server`)

`MEDIA_SERVER_ADMIN_PASS` : set admin user password (will be released later)

`MEDIA_SERVER_DATA_ROUTE_NAME` : set api handler name (example: `/data/`)

`MEDIA_SERVER_DATA_ROUTE_STORAGE_ROUTE` : set folder destination (example: `/`)

## JWT
Generation `jwt-token` can be initted. This version is not usgin add, upadate and remove files from `data` directory with using API.