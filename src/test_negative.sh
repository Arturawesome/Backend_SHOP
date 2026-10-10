#!/usr/bin/env bash

set -euo pipefail

BASE_URL="http://localhost:4000/api/v1"

if ! command -v curl >/dev/null 2>&1; then
    echo "ERROR: curl is not installed"
    exit 1
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

LAST_BODY="$TMP_DIR/body"
LAST_HEADERS="$TMP_DIR/headers"

# Генерируем UUID, которого практически наверняка нет в БД.
if command -v uuidgen >/dev/null 2>&1; then
    MISSING_CLIENT_ID="$(uuidgen | tr '[:upper:]' '[:lower:]')"
    MISSING_PRODUCT_ID="$(uuidgen | tr '[:upper:]' '[:lower:]')"
    MISSING_SUPPLIER_ID="$(uuidgen | tr '[:upper:]' '[:lower:]')"
    MISSING_IMAGE_ID="$(uuidgen | tr '[:upper:]' '[:lower:]')"
else
    MISSING_CLIENT_ID="10000000-0000-4000-8000-000000000001"
    MISSING_PRODUCT_ID="20000000-0000-4000-8000-000000000002"
    MISSING_SUPPLIER_ID="30000000-0000-4000-8000-000000000003"
    MISSING_IMAGE_ID="40000000-0000-4000-8000-000000000004"
fi

request() {
    local expected_status="$1"
    shift

    : > "$LAST_BODY"
    : > "$LAST_HEADERS"

    local status
    status="$(
        curl -sS \
            -D "$LAST_HEADERS" \
            -o "$LAST_BODY" \
            -w "%{http_code}" \
            "$@"
    )"

    if [[ "$status" != "$expected_status" ]]; then
        echo
        echo "FAILED: expected HTTP $expected_status, got HTTP $status"
        echo "Response headers:"
        cat "$LAST_HEADERS"
        echo
        echo "Response body:"
        cat "$LAST_BODY"
        echo
        exit 1
    fi

    echo "OK: HTTP $status"

    if [[ -s "$LAST_BODY" ]]; then
        echo "Response:"
        cat "$LAST_BODY"
        echo
    fi
}

echo "============================================================"
echo " Appliance Shop API - negative/error tests"
echo " Base URL: $BASE_URL"
echo "============================================================"
echo

# ============================================================
# CLIENTS
# ============================================================

echo "1. DELETE nonexistent client -> 404"
request 404 \
    -X DELETE "$BASE_URL/clients/$MISSING_CLIENT_ID"
echo

echo "2. PATCH address of nonexistent client -> 404"
request 404 \
    -X PATCH "$BASE_URL/clients/$MISSING_CLIENT_ID/address" \
    -H "Content-Type: Application/json" \
    -d '{
        "country": "Russia",
        "city": "Moscow",
        "street": "Test street"
    }'
echo

echo "3. DELETE client with invalid UUID -> 400"
request 400 \
    -X DELETE "$BASE_URL/clients/not-a-uuid"
echo

echo "4. Client search without required name/surname -> 400"
request 400 \
    "$BASE_URL/clients/search"
echo

echo "5. Client list with negative limit -> 400"
request 400 \
    "$BASE_URL/clients?limit=-1&offset=0"
echo

echo "6. Client list with non-integer offset -> 400"
request 400 \
    "$BASE_URL/clients?limit=10&offset=abc"
echo

echo "7. Add client with invalid body -> 400"
request 400 \
    -X POST "$BASE_URL/clients" \
    -H "Content-Type: Application/json" \
    -d '{
        "name": "",
        "surname": "",
        "birthday": "1998-05-12T00:00:00Z",
        "gender": "",
        "address": {
            "country": "",
            "city": "",
            "street": ""
        }
    }'
echo

echo "8. Add client with unknown JSON field -> 400"
request 400 \
    -X POST "$BASE_URL/clients" \
    -H "Content-Type: Application/json" \
    -d '{
        "name": "Test",
        "surname": "Client",
        "birthday": "1998-05-12T00:00:00Z",
        "gender": "male",
        "unknown_field": "must fail",
        "address": {
            "country": "Russia",
            "city": "Moscow",
            "street": "Test"
        }
    }'
echo

# ============================================================
# PRODUCTS
# ============================================================

echo "9. GET nonexistent product -> 404"
request 404 \
    "$BASE_URL/products/$MISSING_PRODUCT_ID"
echo

echo "10. DELETE nonexistent product -> 404"
request 404 \
    -X DELETE "$BASE_URL/products/$MISSING_PRODUCT_ID"
echo

echo "11. PATCH stock of nonexistent product -> 404"
request 404 \
    -X PATCH "$BASE_URL/products/$MISSING_PRODUCT_ID/stock?num=1"
echo

echo "12. PATCH stock without num -> 400"
request 400 \
    -X PATCH "$BASE_URL/products/$MISSING_PRODUCT_ID/stock"
echo

echo "13. PATCH stock with non-integer num -> 400"
request 400 \
    -X PATCH "$BASE_URL/products/$MISSING_PRODUCT_ID/stock?num=abc"
echo

echo "14. PATCH stock with num=0 -> 400"
request 400 \
    -X PATCH "$BASE_URL/products/$MISSING_PRODUCT_ID/stock?num=0"
echo

echo "15. Add product with nonexistent supplier -> 400"
request 400 \
    -X POST "$BASE_URL/products" \
    -H "Content-Type: Application/json" \
    -d "{
        \"name\": \"Negative test product\",
        \"category\": \"test\",
        \"price\": 1000,
        \"supplier_id\": \"$MISSING_SUPPLIER_ID\",
        \"available_stock\": 1
    }"
echo

echo "16. Add product with invalid data -> 400"
request 400 \
    -X POST "$BASE_URL/products" \
    -H "Content-Type: Application/json" \
    -d "{
        \"name\": \"\",
        \"category\": \"\",
        \"price\": -10,
        \"supplier_id\": \"$MISSING_SUPPLIER_ID\",
        \"available_stock\": -1
    }"
echo

# ============================================================
# SUPPLIERS
# ============================================================

echo "17. GET nonexistent supplier -> 404"
request 404 \
    "$BASE_URL/suppliers/$MISSING_SUPPLIER_ID"
echo

echo "18. DELETE nonexistent supplier -> 404"
request 404 \
    -X DELETE "$BASE_URL/suppliers/$MISSING_SUPPLIER_ID"
echo

echo "19. PATCH address of nonexistent supplier -> 404"
request 404 \
    -X PATCH "$BASE_URL/suppliers/$MISSING_SUPPLIER_ID/address" \
    -H "Content-Type: Application/json" \
    -d '{
        "country": "Russia",
        "city": "Moscow",
        "street": "Test street"
    }'
echo

echo "20. Add supplier with invalid body -> 400"
request 400 \
    -X POST "$BASE_URL/suppliers" \
    -H "Content-Type: Application/json" \
    -d '{
        "name": "",
        "phone_number": "",
        "address": {
            "country": "",
            "city": "",
            "street": ""
        }
    }'
echo

# ============================================================
# IMAGES
# ============================================================

echo "21. GET nonexistent image -> 404"
request 404 \
    "$BASE_URL/images/$MISSING_IMAGE_ID"
echo

echo "22. GET image of nonexistent product -> 404"
request 404 \
    "$BASE_URL/products/$MISSING_PRODUCT_ID/image"
echo

echo "23. DELETE nonexistent image -> 404"
request 404 \
    -X DELETE "$BASE_URL/images/$MISSING_IMAGE_ID"
echo

echo "24. PUT nonexistent image -> 404"
request 404 \
    -X PUT "$BASE_URL/images/$MISSING_IMAGE_ID" \
    -H "Content-Type: Application/json" \
    -d '{
        "image": "aGVsbG8="
    }'
echo

echo "25. POST image to nonexistent product -> 400"
request 400 \
    -X POST "$BASE_URL/products/$MISSING_PRODUCT_ID/image" \
    -H "Content-Type: Application/json" \
    -d '{
        "image": "aGVsbG8="
    }'
echo

echo "26. POST empty image -> 400"
request 400 \
    -X POST "$BASE_URL/products/$MISSING_PRODUCT_ID/image" \
    -H "Content-Type: Application/json" \
    -d '{
        "image": ""
    }'
echo

echo "27. PUT image with invalid UUID -> 400"
request 400 \
    -X PUT "$BASE_URL/images/not-a-uuid" \
    -H "Content-Type: Application/json" \
    -d '{
        "image": "aGVsbG8="
    }'
echo

echo "============================================================"
echo " ALL NEGATIVE TESTS PASSED"
echo "============================================================"
