#!/bin/bash

BASE_URL="https://shop.test"
CURL="curl -k -s"

echo "========================================"
echo "       NGINX FINAL TEST"
echo "========================================"
echo


echo "1. DNS: shop.test -> 127.0.0.1"
echo "----------------------------------------"

IP=$(dig +short shop.test | tail -n 1)

if [ "$IP" = "127.0.0.1" ]; then
    echo "OK: shop.test -> $IP"
else
    echo "FAIL: shop.test -> $IP"
fi

echo


echo "2. HTTPS"
echo "----------------------------------------"

STATUS=$($CURL -o /dev/null -w "%{http_code}" "$BASE_URL/")

if [ "$STATUS" = "200" ]; then
    echo "OK: HTTPS works (HTTP $STATUS)"
else
    echo "FAIL: HTTP status = $STATUS"
fi

echo


echo "3. SSL certificate"
echo "----------------------------------------"

CERT_SUBJECT=$(
    openssl s_client \
        -connect shop.test:443 \
        -servername shop.test \
        </dev/null 2>/dev/null \
        | openssl x509 -noout -subject 2>/dev/null
)

echo "$CERT_SUBJECT"

if echo "$CERT_SUBJECT" | grep -q "shop.test"; then
    echo "OK: certificate belongs to shop.test"
else
    echo "FAIL: unexpected certificate"
fi

echo


echo "4. Static files"
echo "----------------------------------------"

for PATH_NAME in "/" "/image.png"; do
    STATUS=$($CURL -o /dev/null -w "%{http_code}" "$BASE_URL$PATH_NAME")

    if [ "$STATUS" = "200" ]; then
        echo "OK: $PATH_NAME -> $STATUS"
    else
        echo "FAIL: $PATH_NAME -> $STATUS"
    fi
done

echo


echo "5. Swagger"
echo "----------------------------------------"

STATUS=$($CURL -L -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/")

if [ "$STATUS" = "200" ]; then
    echo "OK: /api/v1/ -> Swagger ($STATUS)"
else
    echo "FAIL: Swagger -> $STATUS"
fi

echo


echo "6. Nginx status"
echo "----------------------------------------"

STATUS=$($CURL -o /dev/null -w "%{http_code}" "$BASE_URL/status")

if [ "$STATUS" = "200" ]; then
    echo "OK: /status -> $STATUS"
    $CURL "$BASE_URL/status"
else
    echo "FAIL: /status -> $STATUS"
fi

echo


echo "7. GET routing"
echo "----------------------------------------"

$CURL -D - -o /dev/null "$BASE_URL/api/products" \
    | grep -Ei "HTTP/|X-Application|X-Upstream-Addr"

echo


echo "8. GET load balancing (2:1:1)"
echo "----------------------------------------"

echo "Sending 40 GET requests..."

for i in {1..40}; do
    $CURL -D - -o /dev/null "$BASE_URL/api/products" \
        | awk -F': ' '
            tolower($1)=="x-upstream-addr" {
                gsub("\r","",$2);
                print $2
            }
        '
done | sort | uniq -c

echo
echo "Expected ratio:"
echo "  :4000 ~ 20"
echo "  :4001 ~ 10"
echo "  :4002 ~ 10"
echo


echo "9. Write routing"
echo "----------------------------------------"

for METHOD in POST PUT PATCH DELETE; do

    echo "$METHOD:"

    $CURL \
        -X "$METHOD" \
        -D - \
        -o /dev/null \
        "$BASE_URL/api/products" \
        | grep -Ei "HTTP/|X-Application|X-Upstream-Addr"

    echo
done

echo "All write requests should use write_Application"
echo "and backend :4000."
echo


echo "10. API must NOT be cached"
echo "----------------------------------------"

HEADERS=$($CURL -D - -o /dev/null "$BASE_URL/api/products")

if echo "$HEADERS" | grep -qi "X-Cache-Status"; then
    echo "FAIL: /api contains X-Cache-Status"
else
    echo "OK: /api is not cached"
fi

echo


echo "11. /admin proxy cache"
echo "----------------------------------------"

echo "First request:"
$CURL -D - -o /dev/null "$BASE_URL/admin/" \
    | grep -Ei "HTTP/|X-Cache-Status"

echo

echo "Second request:"
$CURL -D - -o /dev/null "$BASE_URL/admin/" \
    | grep -Ei "HTTP/|X-Cache-Status"

echo
echo "Expected: first MISS, second HIT."
echo "If the cache already contains the page,"
echo "both requests may show HIT."
echo


echo "12. GZIP: text content"
echo "----------------------------------------"

GZIP_HEADERS=$(
    $CURL \
        -H "Accept-Encoding: gzip" \
        -D - \
        -o /dev/null \
        "$BASE_URL/api/v1/index.html"
)

echo "$GZIP_HEADERS" \
    | grep -Ei "HTTP/|Content-Type|Content-Encoding"

if echo "$GZIP_HEADERS" | grep -qi "Content-Encoding: gzip"; then
    echo "OK: text is gzip compressed"
else
    echo "WARNING: gzip not detected"
fi

echo


echo "13. GZIP: PNG must NOT be compressed"
echo "----------------------------------------"

PNG_HEADERS=$(
    $CURL \
        -H "Accept-Encoding: gzip" \
        -D - \
        -o /dev/null \
        "$BASE_URL/image.png"
)

echo "$PNG_HEADERS" \
    | grep -Ei "HTTP/|Content-Type|Content-Encoding"

if echo "$PNG_HEADERS" | grep -qi "Content-Encoding: gzip"; then
    echo "FAIL: PNG is gzip compressed"
else
    echo "OK: PNG is not gzip compressed"
fi

echo


echo "========================================"
echo "          TEST FINISHED"
echo "========================================"