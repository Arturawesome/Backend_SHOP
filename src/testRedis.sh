#!/usr/bin/env bash

set -euo pipefail

BASE_URL="http://localhost:4000/api/v1"

EMAIL="seed@example.com"
PASSWORD="password123"

FIRST_IMAGE="/tmp/redis_image_first.bin"
SECOND_IMAGE="/tmp/redis_image_second.bin"
CHANGED_IMAGE="/tmp/redis_image_changed.bin"

TEST_IMAGE="./static/images4.png"

echo "========================================"
echo " Redis image cache test"
echo "========================================"

command -v curl >/dev/null 2>&1 || {
    echo "curl is required"
    exit 1
}

command -v jq >/dev/null 2>&1 || {
    echo "jq is required"
    exit 1
}

command -v docker >/dev/null 2>&1 || {
    echo "docker is required"
    exit 1
}

[[ -f "$TEST_IMAGE" ]] || {
    echo "Test image not found: $TEST_IMAGE"
    exit 1
}

# ------------------------------------------------------------
# 1. Login
# ------------------------------------------------------------

echo
echo "[1] Login"

LOGIN_RESPONSE=$(curl -sS \
    -X POST \
    "$BASE_URL/auth" \
    -H "Content-Type: application/json" \
    -d "{
        \"email\": \"$EMAIL\",
        \"password\": \"$PASSWORD\"
    }"
)

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')

if [[ -z "$TOKEN" || "$TOKEN" == "null" ]]; then
    echo "Cannot obtain JWT"
    echo "$LOGIN_RESPONSE"
    exit 1
fi

echo "JWT received"

# ------------------------------------------------------------
# 2. Get products and select one with image
# ------------------------------------------------------------

echo
echo "[2] Find product with image"

PRODUCTS=$(curl -sS \
    "$BASE_URL/products" \
    -H "Authorization: Bearer $TOKEN"
)

PRODUCT=$(echo "$PRODUCTS" | jq -c \
    '[.[] | select(.image_id != null)][0]')

if [[ "$PRODUCT" == "null" || -z "$PRODUCT" ]]; then
    echo "No product with image found."
    echo "Run createDB.sh first."
    exit 1
fi

PRODUCT_ID=$(echo "$PRODUCT" | jq -r '.id')
PRODUCT_NAME=$(echo "$PRODUCT" | jq -r '.name')
IMAGE_ID=$(echo "$PRODUCT" | jq -r '.image_id')

echo "Product:  $PRODUCT_NAME"
echo "Product ID: $PRODUCT_ID"
echo "Image ID:   $IMAGE_ID"

REDIS_KEY="image:$IMAGE_ID"

# ------------------------------------------------------------
# 3. Clear cache
# ------------------------------------------------------------

echo
echo "[3] Clear Redis key"

docker compose exec -T redis \
    redis-cli DEL "$REDIS_KEY" >/dev/null

EXISTS=$(docker compose exec -T redis \
    redis-cli EXISTS "$REDIS_KEY" | tr -d '\r')

if [[ "$EXISTS" != "0" ]]; then
    echo "Failed to clear Redis key"
    exit 1
fi

echo "Redis key removed"
echo
echo "Watch Gateway console now."
echo "Next request should print:"
echo "IMAGE FROM POSTGRES"

# ------------------------------------------------------------
# 4. First request -> PostgreSQL
# ------------------------------------------------------------

echo
echo "[4] First GET — expected CACHE MISS / POSTGRES"

STATUS=$(curl -sS \
    -o "$FIRST_IMAGE" \
    -w "%{http_code}" \
    "$BASE_URL/images/$IMAGE_ID" \
    -H "Authorization: Bearer $TOKEN"
)

if [[ "$STATUS" != "200" ]]; then
    echo "Expected HTTP 200, got $STATUS"
    exit 1
fi

if [[ ! -s "$FIRST_IMAGE" ]]; then
    echo "Downloaded image is empty"
    exit 1
fi

echo "First image downloaded: $(wc -c < "$FIRST_IMAGE") bytes"

# ------------------------------------------------------------
# 5. Check Redis
# ------------------------------------------------------------

echo
echo "[5] Check Redis"

EXISTS=$(docker compose exec -T redis \
    redis-cli EXISTS "$REDIS_KEY" | tr -d '\r')

if [[ "$EXISTS" != "1" ]]; then
    echo "FAIL: image was not written to Redis"
    exit 1
fi

echo "Redis key exists: $REDIS_KEY"

REDIS_SIZE=$(docker compose exec -T redis \
    redis-cli STRLEN "$REDIS_KEY" | tr -d '\r')

echo "Cached size: $REDIS_SIZE bytes"

# ------------------------------------------------------------
# 6. Second request -> Redis
# ------------------------------------------------------------

echo
echo "[6] Second GET — expected CACHE HIT"
echo
echo "Gateway console should now print:"
echo "IMAGE FROM CACHE"

STATUS=$(curl -sS \
    -o "$SECOND_IMAGE" \
    -w "%{http_code}" \
    "$BASE_URL/images/$IMAGE_ID" \
    -H "Authorization: Bearer $TOKEN"
)

if [[ "$STATUS" != "200" ]]; then
    echo "Expected HTTP 200, got $STATUS"
    exit 1
fi

# ------------------------------------------------------------
# 7. Compare results
# ------------------------------------------------------------

echo
echo "[7] Compare PostgreSQL and Redis responses"

if cmp -s "$FIRST_IMAGE" "$SECOND_IMAGE"; then
    echo "OK: images are identical"
else
    echo "FAIL: cached image differs from PostgreSQL image"
    exit 1
fi

# ------------------------------------------------------------
# 8. Authorization check
# ------------------------------------------------------------

echo
echo "[8] Check authorization"

STATUS=$(curl -sS \
    -o /dev/null \
    -w "%{http_code}" \
    "$BASE_URL/images/$IMAGE_ID"
)

if [[ "$STATUS" == "401" ]]; then
    echo "OK: request without JWT returns 401"
else
    echo "FAIL: expected 401, got $STATUS"
    exit 1
fi

# ------------------------------------------------------------
# 9. Cache invalidation after PUT
# ------------------------------------------------------------

echo
echo "[9] Test cache invalidation after image update"

NEW_IMAGE_BASE64=$(base64 < "$TEST_IMAGE" | tr -d '\n')

STATUS=$(curl -sS \
    -o /dev/null \
    -w "%{http_code}" \
    -X PUT \
    "$BASE_URL/images/$IMAGE_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"image\": \"$NEW_IMAGE_BASE64\"
    }"
)

if [[ "$STATUS" != "204" ]]; then
    echo "FAIL: PUT expected 204, got $STATUS"
    exit 1
fi

echo "Image updated in PostgreSQL"

EXISTS=$(docker compose exec -T redis \
    redis-cli EXISTS "$REDIS_KEY" | tr -d '\r')

if [[ "$EXISTS" == "0" ]]; then
    echo "OK: old cache entry was invalidated"
else
    echo "FAIL: stale image is still stored in Redis"
    echo "ChangeImage() must delete/update Redis key"
    exit 1
fi

echo
echo "Request updated image..."

STATUS=$(curl -sS \
    -o "$CHANGED_IMAGE" \
    -w "%{http_code}" \
    "$BASE_URL/images/$IMAGE_ID" \
    -H "Authorization: Bearer $TOKEN"
)

if [[ "$STATUS" != "200" ]]; then
    echo "FAIL: expected 200, got $STATUS"
    exit 1
fi

if cmp -s "$CHANGED_IMAGE" "$TEST_IMAGE"; then
    echo "OK: updated image returned correctly"
else
    echo "FAIL: response contains stale image"
    exit 1
fi

echo
echo "========================================"
echo " ALL REDIS CACHE TESTS PASSED"
echo "========================================"
echo
echo "Expected Gateway log sequence:"
echo
echo "  IMAGE FROM POSTGRES"
echo "  IMAGE FROM CACHE"
echo "  IMAGE FROM POSTGRES    <- after PUT"
echo