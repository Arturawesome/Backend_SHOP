#!/usr/bin/env bash

set -euo pipefail

BASE_URL="http://localhost:4000/api/v1"
IMAGE_FILE="./testImage.png"

# Нужен jq, потому что POST-методы создания у текущего API
# не возвращают созданный UUID в response body.
if ! command -v curl >/dev/null 2>&1; then
    echo "ERROR: curl is not installed"
    exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
    echo "ERROR: jq is not installed"
    echo "macOS: brew install jq"
    exit 1
fi

if [[ ! -f "$IMAGE_FILE" ]]; then
    echo "ERROR: $IMAGE_FILE not found"
    echo "Put testImage.png in the project root and run this script from there."
    exit 1
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

RUN_ID="$(date +%s)"
CLIENT_NAME="ArturTest${RUN_ID}"
CLIENT_SURNAME="NasyrovTest${RUN_ID}"
SUPPLIER_NAME="SupplierTest${RUN_ID}"
PRODUCT_NAME="ProductTest${RUN_ID}"

IMAGE_BASE64="$(base64 < "$IMAGE_FILE" | tr -d '\n')"

LAST_BODY="$TMP_DIR/body"
LAST_HEADERS="$TMP_DIR/headers"

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
}

print_json_body() {
    if [[ -s "$LAST_BODY" ]]; then
        jq . "$LAST_BODY"
    fi
}

echo "============================================================"
echo " Appliance Shop API - handler smoke tests"
echo " Base URL: $BASE_URL"
echo "============================================================"
echo

# ============================================================
# SUPPLIERS
# ============================================================

echo "1/20 POST /suppliers - SupplierAdd"
request 201 \
    -X POST "$BASE_URL/suppliers" \
    -H "Content-Type: Application/json" \
    -d "{
        \"name\": \"$SUPPLIER_NAME\",
        \"phone_number\": \"+74951234567\",
        \"address\": {
            \"country\": \"Russia\",
            \"city\": \"Moscow\",
            \"street\": \"Baumanskaya street\"
        }
    }"
echo

echo "2/20 GET /suppliers - SupplierList"
request 200 \
    "$BASE_URL/suppliers"
print_json_body

SUPPLIER_ID="$(
    jq -r --arg name "$SUPPLIER_NAME" \
        '[.[] | select(.name == $name)][-1].id // empty' \
        "$LAST_BODY"
)"

if [[ -z "$SUPPLIER_ID" ]]; then
    echo "FAILED: cannot find created supplier in SupplierList response"
    exit 1
fi

echo "Created supplier ID: $SUPPLIER_ID"
echo

echo "3/20 GET /suppliers/{id} - SupplierGetByID"
request 200 \
    "$BASE_URL/suppliers/$SUPPLIER_ID"
print_json_body
echo

echo "4/20 PATCH /suppliers/{id}/address - SupplierPatchAddress"
request 204 \
    -X PATCH "$BASE_URL/suppliers/$SUPPLIER_ID/address" \
    -H "Content-Type: Application/json" \
    -d '{
        "country": "Germany",
        "city": "Berlin",
        "street": "Teststrasse 10"
    }'
echo

# ============================================================
# PRODUCTS
# ============================================================

echo "5/20 POST /products - ProductAdd"
request 201 \
    -X POST "$BASE_URL/products" \
    -H "Content-Type: Application/json" \
    -d "{
        \"name\": \"$PRODUCT_NAME\",
        \"category\": \"washing_machine\",
        \"price\": 59990,
        \"supplier_id\": \"$SUPPLIER_ID\",
        \"available_stock\": 5
    }"
echo

echo "6/20 GET /products - ProductGetAvailable"
request 200 \
    "$BASE_URL/products"
print_json_body

PRODUCT_ID="$(
    jq -r --arg name "$PRODUCT_NAME" \
        '[.[] | select(.name == $name)][-1].id // empty' \
        "$LAST_BODY"
)"

if [[ -z "$PRODUCT_ID" ]]; then
    echo "FAILED: cannot find created product in ProductGetAvailable response"
    exit 1
fi

echo "Created product ID: $PRODUCT_ID"
echo

echo "7/20 GET /products/{id} - ProductGetByID"
request 200 \
    "$BASE_URL/products/$PRODUCT_ID"
print_json_body
echo

echo "8/20 PATCH /products/{id}/stock?num=1 - ProductDecrease"
request 204 \
    -X PATCH "$BASE_URL/products/$PRODUCT_ID/stock?num=1"
echo

# Проверяем, что available_stock реально уменьшился с 5 до 4.
echo "Check product stock after ProductDecrease"
request 200 \
    "$BASE_URL/products/$PRODUCT_ID"

STOCK="$(jq -r '.available_stock' "$LAST_BODY")"
if [[ "$STOCK" != "4" ]]; then
    echo "FAILED: expected available_stock=4, got $STOCK"
    exit 1
fi
echo "OK: available_stock = $STOCK"
echo

# ============================================================
# IMAGES
# ============================================================

echo "9/20 POST /products/{id}/image - ImageAdd"
request 201 \
    -X POST "$BASE_URL/products/$PRODUCT_ID/image" \
    -H "Content-Type: Application/json" \
    -d "$(jq -n --arg image "$IMAGE_BASE64" '{image: $image}')"
echo

# После добавления изображения узнаём image_id через ProductGetByID.
request 200 \
    "$BASE_URL/products/$PRODUCT_ID"

IMAGE_ID="$(jq -r '.image_id // empty' "$LAST_BODY")"

if [[ -z "$IMAGE_ID" || "$IMAGE_ID" == "null" ]]; then
    echo "FAILED: product.image_id is empty after ImageAdd"
    exit 1
fi

echo "Created image ID: $IMAGE_ID"
echo

echo "10/20 GET /products/{id}/image - ImageGetByProductID"
request 200 \
    "$BASE_URL/products/$PRODUCT_ID/image"

CONTENT_TYPE="$(awk 'BEGIN{IGNORECASE=1} /^Content-Type:/ {gsub(/\r/,""); print $2}' "$LAST_HEADERS" | tail -n1)"
CONTENT_DISPOSITION="$(grep -i '^Content-Disposition:' "$LAST_HEADERS" | tr -d '\r' | tail -n1 || true)"

if [[ "$CONTENT_TYPE" != "Application/octet-stream" ]]; then
    echo "FAILED: expected Content-Type Application/octet-stream, got: $CONTENT_TYPE"
    exit 1
fi

if [[ "$CONTENT_DISPOSITION" != *"attachment"* ]]; then
    echo "FAILED: Content-Disposition does not contain attachment"
    echo "$CONTENT_DISPOSITION"
    exit 1
fi

cp "$LAST_BODY" "$TMP_DIR/image-by-product.bin"
echo "OK: binary response, Content-Type=$CONTENT_TYPE"
echo "OK: $CONTENT_DISPOSITION"
echo

echo "11/20 GET /images/{id} - ImageGet"
request 200 \
    "$BASE_URL/images/$IMAGE_ID"

CONTENT_TYPE="$(awk 'BEGIN{IGNORECASE=1} /^Content-Type:/ {gsub(/\r/,""); print $2}' "$LAST_HEADERS" | tail -n1)"
if [[ "$CONTENT_TYPE" != "Application/octet-stream" ]]; then
    echo "FAILED: expected Content-Type Application/octet-stream, got: $CONTENT_TYPE"
    exit 1
fi

cp "$LAST_BODY" "$TMP_DIR/image-by-id.bin"
echo "OK: binary response, Content-Type=$CONTENT_TYPE"
echo

echo "12/20 PUT /images/{id} - ImageChange"
request 204 \
    -X PUT "$BASE_URL/images/$IMAGE_ID" \
    -H "Content-Type: Application/json" \
    -d "$(jq -n --arg image "$IMAGE_BASE64" '{image: $image}')"
echo

# ============================================================
# CLIENTS
# ============================================================

echo "13/20 POST /clients - ClientAdd"
request 201 \
    -X POST "$BASE_URL/clients" \
    -H "Content-Type: Application/json" \
    -d "{
        \"name\": \"$CLIENT_NAME\",
        \"surname\": \"$CLIENT_SURNAME\",
        \"birthday\": \"1998-05-12T00:00:00Z\",
        \"gender\": \"male\",
        \"address\": {
            \"country\": \"Russia\",
            \"city\": \"Moscow\",
            \"street\": \"Baumanskaya street\"
        }
    }"
echo

echo "14/20 GET /clients/search - ClientSearch"
request 200 \
    "$BASE_URL/clients/search?name=$CLIENT_NAME&surname=$CLIENT_SURNAME"
print_json_body

CLIENT_ID="$(
    jq -r --arg name "$CLIENT_NAME" --arg surname "$CLIENT_SURNAME" \
        '[.[] | select(.name == $name and .surname == $surname)][-1].id // empty' \
        "$LAST_BODY"
)"

if [[ -z "$CLIENT_ID" ]]; then
    echo "FAILED: cannot find created client in ClientSearch response"
    exit 1
fi

echo "Created client ID: $CLIENT_ID"
echo

echo "15/20 GET /clients?limit=10&offset=0 - ClientList"
request 200 \
    "$BASE_URL/clients?limit=10&offset=0"
print_json_body
echo

echo "16/20 PATCH /clients/{id}/address - ClientPatchAddress"
request 204 \
    -X PATCH "$BASE_URL/clients/$CLIENT_ID/address" \
    -H "Content-Type: Application/json" \
    -d '{
        "country": "Germany",
        "city": "Munich",
        "street": "Clientstrasse 20"
    }'
echo

# ============================================================
# DELETE / CLEANUP
# ============================================================

echo "17/20 DELETE /images/{id} - ImageRemove"
request 204 \
    -X DELETE "$BASE_URL/images/$IMAGE_ID"
echo

echo "18/20 DELETE /products/{id} - ProductRemoveByID"
request 204 \
    -X DELETE "$BASE_URL/products/$PRODUCT_ID"
echo

echo "19/20 DELETE /suppliers/{id} - SupplierRemoveByID"
request 204 \
    -X DELETE "$BASE_URL/suppliers/$SUPPLIER_ID"
echo

echo "20/20 DELETE /clients/{id} - ClientRemove"
request 204 \
    -X DELETE "$BASE_URL/clients/$CLIENT_ID"
echo

echo "============================================================"
echo " ALL 20 HANDLER TESTS PASSED"
echo "============================================================"
