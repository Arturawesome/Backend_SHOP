# NGINX

## 1. Basic Nginx structure

Basic `nginx.conf` structure:

```nginx
events {
}

http {
    # Global HTTP configuration:
    # upstreams, maps, cache, gzip, etc.

    server {
        # Virtual server configuration:
        # listen, server_name, locations, proxying, etc.
    }
}
```

Check where Homebrew installed Nginx:

```bash
brew --prefix nginx
```

On Apple Silicon macOS, the configuration is usually located at:

```text
/opt/homebrew/etc/nginx/nginx.conf
```

Check Nginx configuration:

```bash
nginx -t
```

Reload Nginx after changing the configuration:

```bash
nginx -s reload
```

---

# 2. Reverse proxy

Nginx can accept a request from a client and forward it to the Go Application.

For example:

```nginx
location /api/ {
    proxy_pass http://127.0.0.1:4000/api/v1/;

    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

Example request flow:

```text
Client:
GET /api/products

        ↓ Nginx

Backend:
GET /api/v1/products
```

### Proxy headers

```nginx
proxy_set_header Host $host;
```

Passes the original hostname.

```nginx
proxy_set_header X-Real-IP $remote_addr;
```

Passes the client's IP address.

```nginx
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
```

Keeps the proxy chain and original client address.

```nginx
proxy_set_header X-Forwarded-Proto $scheme;
```

Tells the backend whether the original request used HTTP or HTTPS.

---

# 3. Upstream and load balancing

`upstream` describes a group of backend servers.

```nginx
upstream get_Application {
    server localhost:4000 weight=2;
    server localhost:4001 weight=1;
    server localhost:4002 weight=1;
}
```

The weights define the request distribution:

```text
4000 → 2 parts
4001 → 1 part
4002 → 1 part
```

Expected ratio:

```text
4000 → 50%
4001 → 25%
4002 → 25%
```

For example, approximately 40 requests should be distributed as:

```text
4000 → 20
4001 → 10
4002 → 10
```

Create a separate upstream for write operations:

```nginx
upstream write_Application {
    server localhost:4000;
}
```

Thus only the main backend receives write requests.

---

# 4. Routing GET and write requests

The `map` directive can select an upstream depending on the HTTP method.

```nginx
map $request_method $Application {
    GET     http://get_Application;
    default http://write_Application;
}
```

The logic is:

```text
GET
 ↓
get_Application
 ↓
4000 / 4001 / 4002


POST / PUT / PATCH / DELETE
 ↓
write_Application
 ↓
4000
```

API location:

```nginx
location /api/ {
    rewrite ^/api/(.*)$ /api/v1/$1 break;

    proxy_pass $Application;

    add_header X-Application $Application always;
    add_header X-Upstream-Addr $upstream_addr always;

    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

`rewrite` transforms:

```text
/api/products
```

into:

```text
/api/v1/products
```

The diagnostic headers:

```text
X-Application
X-Upstream-Addr
```

show which upstream group and backend handled the request.

---

# 5. Swagger routing

Redirect `/api/v1` to `/api/v1/`:

```nginx
location = /api/v1 {
    return 302 /api/v1/;
}
```

`=` means an **exact location match**.

Serve the Swagger index:

```nginx
location = /api/v1/ {
    proxy_pass http://127.0.0.1:4000/swagger/index.html;

    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

Serve Swagger resources:

```nginx
location /api/v1/ {
    proxy_pass http://127.0.0.1:4000/swagger/;

    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

Result:

```text
https://shop.test/api/v1/
        ↓
Swagger UI
```

---

# 6. Static files

Nginx can serve static files without sending the request to the Go backend.

```nginx
location / {
    root /Users/arturawesome/School21Projects/BE3_Nginx_Pro_ID_1269323-1/src/static;

    index index.html;

    try_files $uri $uri/ =404;

    expires 10m;
}
```

For example:

```text
/
    → static/index.html

/image.png
    → static/image.png
```

`try_files` checks whether the requested file exists and returns `404` otherwise.

---

# 7. pgAdmin proxy

Proxy `/admin/` to pgAdmin:

```nginx
location /admin/ {
    proxy_pass http://127.0.0.1:5050/;

    proxy_cache static_cache;
    proxy_cache_bypass $skip_cache;
    proxy_no_cache $skip_cache;
    proxy_cache_valid 200 10m;

    add_header X-Cache-Status $upstream_cache_status always;

    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

Request flow:

```text
https://shop.test/admin/
        ↓
Nginx
        ↓
localhost:5050
        ↓
pgAdmin
```

> **Note:** Caching authenticated pgAdmin pages is used here only to demonstrate `proxy_cache` for the educational assignment. It should generally not be enabled this way in a production environment.

---

# 8. Nginx status

Enable the built-in Nginx status page:

```nginx
location = /status {
    stub_status;
}
```

Check:

```bash
curl -k https://shop.test/status
```

Example output:

```text
Active connections: 1
server accepts handled requests
...
Reading: 0 Writing: 1 Waiting: 0
```

---

# 9. Cache

Create a directory for the Nginx cache:

```bash
mkdir -p \
/Users/arturawesome/School21Projects/BE3_Nginx_Pro_ID_1269323-1/src/cache/nginx
```

Configure the cache inside the `http` block:

```nginx
proxy_cache_path /Users/arturawesome/School21Projects/BE3_Nginx_Pro_ID_1269323-1/src/cache/nginx
                 levels=1:2
                 keys_zone=static_cache:32m
                 inactive=60m
                 max_size=2g;
```

Configure caching only for GET requests:

```nginx
map $request_method $skip_cache {
    GET     0;
    default 1;
}
```

Use the cache in a proxied location:

```nginx
proxy_cache static_cache;

proxy_cache_bypass $skip_cache;
proxy_no_cache $skip_cache;

proxy_cache_valid 200 10m;

add_header X-Cache-Status $upstream_cache_status always;
```

Typical cache states:

```text
MISS → response was not in cache
HIT  → response was returned from cache
```

The `/api/` location does **not** contain `proxy_cache`, therefore API requests are not cached.

For local static files:

```nginx
expires 10m;
```

can be used to send client-side cache headers.

---

# 10. Gzip compression

Enable gzip inside the `http` block:

```nginx
gzip on;

gzip_comp_level 5;
gzip_min_length 256;
gzip_proxied any;

gzip_types
    text/plain
    text/css
    text/xml
    Application/json
    Application/javascript
    Application/xml
    Application/xml+rss
    Application/xhtml+xml
    image/svg+xml;
```

Text formats are compressed:

```text
HTML
CSS
JavaScript
JSON
XML
SVG
```

Media formats are intentionally not included:

```text
image/png
image/jpeg
image/gif
image/webp
video/*
```

PNG and JPEG are already compressed formats, so additional gzip compression is generally unnecessary.

### Check gzip

```bash
curl -k -s -D - -o /dev/null \
    -H "Accept-Encoding: gzip" \
    https://shop.test/api/v1/index.html
```

Expected header:

```text
Content-Encoding: gzip
```

### Check that PNG is not compressed

```bash
curl -k -s -D - -o /dev/null \
    -H "Accept-Encoding: gzip" \
    https://shop.test/image.png
```

There should be:

```text
Content-Type: image/png
```

but no:

```text
Content-Encoding: gzip
```

---

# DNS

## 11. Install dnsmasq

Install `dnsmasq`:

```bash
brew install dnsmasq
```

Edit the configuration:

```bash
nano /opt/homebrew/etc/dnsmasq.conf
```

Add:

```text
address=/shop.test/127.0.0.1
```

This creates the mapping:

```text
shop.test → 127.0.0.1
```

Start `dnsmasq`:

```bash
sudo brew services start dnsmasq
```

After changing its configuration:

```bash
sudo brew services restart dnsmasq
```

Check `dnsmasq` directly:

```bash
dig @127.0.0.1 shop.test
```

Expected address:

```text
127.0.0.1
```

---

# 12. macOS DNS resolver

`dnsmasq` knows about `shop.test`, but macOS must also know that `.test` queries should be sent to `dnsmasq`.

Create the resolver directory:

```bash
sudo mkdir -p /etc/resolver
```

Create the resolver:

```bash
sudo nano /etc/resolver/test
```

Add:

```text
nameserver 127.0.0.1
```

Now macOS sends requests for the `.test` domain to the DNS server running on `127.0.0.1`.

Check all macOS resolvers:

```bash
scutil --dns
```

There should be a resolver similar to:

```text
domain   : test
nameserver[0] : 127.0.0.1
```

Clear the DNS cache if necessary:

```bash
sudo dscacheutil -flushcache
sudo killall -HUP mDNSResponder
```

Check:

```bash
ping -c 3 shop.test
```

Expected:

```text
shop.test (127.0.0.1)
```

The complete DNS chain:

```text
shop.test
    ↓
macOS resolver for .test
    ↓
127.0.0.1:53
    ↓
dnsmasq
    ↓
127.0.0.1
```

---

# HTTPS

## 13. Generate SSL certificate and private key

Create a directory:

```bash
mkdir -p certs
```

Generate a self-signed certificate:

```bash
openssl req \
    -x509 \
    -nodes \
    -days 365 \
    -newkey rsa:2048 \
    -keyout certs/shop.test.key \
    -out certs/shop.test.crt \
    -subj "/CN=shop.test" \
    -addext "subjectAltName=DNS:shop.test"
```

Generated files:

```text
shop.test.key → private key
shop.test.crt → certificate
```

`CN` specifies the certificate name:

```text
CN=shop.test
```

`subjectAltName` explicitly declares that the certificate is valid for:

```text
DNS:shop.test
```

Check the certificate:

```bash
openssl x509 \
    -in certs/shop.test.crt \
    -text \
    -noout
```

Look for:

```text
Subject: CN=shop.test

X509v3 Subject Alternative Name:
    DNS:shop.test
```

---

# 14. Enable HTTPS in Nginx

HTTPS settings must be placed **inside a `server` block inside `http`**:

```nginx
http {

    # ...

    server {
        listen 443 ssl;
        server_name shop.test;

        ssl_certificate /Users/arturawesome/School21Projects/BE3_Nginx_Pro_ID_1269323-1/src/certs/shop.test.crt;

        ssl_certificate_key /Users/arturawesome/School21Projects/BE3_Nginx_Pro_ID_1269323-1/src/certs/shop.test.key;

        # locations...
    }
}
```

`listen 443 ssl` means:

```text
443 → standard HTTPS port
ssl → enable TLS for this server
```

`server_name` connects this virtual server with the local domain:

```text
shop.test
```

---

# 15. Test HTTPS

Because the certificate is self-signed, `curl` does not trust it by default.

Use:

```bash
curl -k https://shop.test/
```

`-k` allows `curl` to accept the self-signed certificate.

Check API:

```bash
curl -k https://shop.test/api/products
```

Check Swagger:

```bash
curl -k -L https://shop.test/api/v1/
```

Check Nginx status:

```bash
curl -k https://shop.test/status
```

Check the certificate served by Nginx:

```bash
openssl s_client \
    -connect shop.test:443 \
    -servername shop.test
```

A shorter certificate check:

```bash
openssl s_client \
    -connect shop.test:443 \
    -servername shop.test \
    </dev/null 2>/dev/null \
    | openssl x509 -noout -subject -issuer -dates
```

For a self-signed certificate, `subject` and `issuer` normally refer to the same certificate identity.

---

# 16. Final architecture

```text
                         Client
                            │
                    https://shop.test
                            │
                            ▼
                     macOS resolver
                            │
                            ▼
                        dnsmasq
                            │
                     127.0.0.1
                            │
                            ▼
                    Nginx :443 TLS
                            │
          ┌─────────────────┼─────────────────┐
          │                 │                 │
          ▼                 ▼                 ▼
          /               /api/            /admin/
          │                 │                 │
       static          HTTP method          pgAdmin
                            │
                    ┌───────┴────────┐
                    │                │
                   GET            WRITE
                    │                │
                    ▼                ▼
            get_Application   write_Application
                    │                │
             ┌──────┼──────┐         │
             ▼      ▼      ▼         ▼
           :4000  :4001  :4002     :4000
             2       1       1
```

Additional functionality:

```text
GET /api/*
    → load balancing 2:1:1
    → NO CACHE

POST / PUT / PATCH / DELETE /api/*
    → main backend :4000

/admin/*
    → pgAdmin
    → proxy_cache

/
    → static files
    → client cache

HTML / CSS / JS / JSON / XML
    → gzip

PNG / JPEG / media
    → no gzip

/status
    → Nginx stub_status

shop.test
    → dnsmasq
    → 127.0.0.1

HTTPS
    → self-signed certificate
    → TLS on port 443
```

---

# 17. Useful commands

Find Nginx installed by Homebrew:

```bash
brew --prefix nginx
```

Check configuration:

```bash
nginx -t
```

Reload configuration:

```bash
nginx -s reload
```

Check Nginx processes:

```bash
ps aux | grep nginx
```

Check listening ports:

```bash
lsof -nP -iTCP -sTCP:LISTEN | grep nginx
```

Check DNS:

```bash
dig @127.0.0.1 shop.test
```

Check macOS DNS configuration:

```bash
scutil --dns
```

Check HTTPS:

```bash
curl -k https://shop.test/
```

Check response headers:

```bash
curl -k -s -D - -o /dev/null https://shop.test/
```

Check configuration and immediately reload Nginx:

```bash
nginx -t && nginx -s reload
```

---

# 18. Complete request flow

For example:

```text
GET https://shop.test/api/products
```

goes through:

```text
1. shop.test
       ↓
2. macOS DNS resolver
       ↓
3. dnsmasq
       ↓
4. 127.0.0.1:443
       ↓
5. TLS termination in Nginx
       ↓
6. location /api/
       ↓
7. rewrite:
   /api/products → /api/v1/products
       ↓
8. map:
   GET → get_Application
       ↓
9. upstream load balancing:
   4000 : 4001 : 4002
     2  :   1  :   1
       ↓
10. Go backend
       ↓
11. response → Nginx → HTTPS → client
```

This is the complete request path from the local domain to the backend Application.