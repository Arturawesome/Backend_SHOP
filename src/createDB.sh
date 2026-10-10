#!/usr/bin/env bash

set -euo pipefail

BASE_URL="http://localhost:4000/api/v1"

IMAGE_DIR="./static"
IMAGE_FILES=(
    "$IMAGE_DIR/images0.png"
    "$IMAGE_DIR/images1.png"
    "$IMAGE_DIR/images2.png"
    "$IMAGE_DIR/images3.png"
    "$IMAGE_DIR/images4.png"
)

SEED_EMAIL="seed@example.com"
SEED_PASSWORD="password123"

# ============================================================
# Helpers
# ============================================================

log() {
    echo
    echo "==> $1"
}

fail() {
    echo "ERROR: $1" >&2
    exit 1
}

check_dependencies() {
    command -v curl >/dev/null 2>&1 || fail "curl is required"
    command -v jq >/dev/null 2>&1 || fail "jq is required"

    for image_file in "${IMAGE_FILES[@]}"; do
        if [[ ! -f "$image_file" ]]; then
            fail "Image file not found: $image_file"
        fi
    done
}

# Выполняет curl и возвращает body.
# HTTP status записывается в глобальную переменную HTTP_STATUS.
request() {
    local method="$1"
    local url="$2"
    local token="${3:-}"
    local body="${4:-}"

    local tmp
    tmp=$(mktemp)

    local args=(
        -sS
        -o "$tmp"
        -w "%{http_code}"
        -X "$method"
        "$url"
    )

    if [[ -n "$token" ]]; then
        args+=(
            -H "Authorization: Bearer $token"
        )
    fi

    if [[ -n "$body" ]]; then
        args+=(
            -H "Content-Type: application/json"
            -d "$body"
        )
    fi

    HTTP_STATUS=$(curl "${args[@]}")
    HTTP_BODY=$(cat "$tmp")

    rm -f "$tmp"
}

expect_status() {
    local expected="$1"

    if [[ "$HTTP_STATUS" != "$expected" ]]; then
        echo "Expected HTTP $expected, got HTTP $HTTP_STATUS"
        echo "Response:"
        echo "$HTTP_BODY"
        exit 1
    fi
}

# ============================================================
# Authentication
# ============================================================

get_token() {
    log "Authenticating seed user"

    # Сначала пробуем login. Это позволяет повторно запускать seed.
    request \
        POST \
        "$BASE_URL/auth" \
        "" \
        "{
            \"email\": \"$SEED_EMAIL\",
            \"password\": \"$SEED_PASSWORD\"
        }"

    if [[ "$HTTP_STATUS" == "200" ]]; then
        TOKEN=$(echo "$HTTP_BODY" | jq -r '.access_token')

        [[ "$TOKEN" != "null" && -n "$TOKEN" ]] \
            || fail "Login response does not contain access_token"

        echo "Seed user already exists"
        return
    fi

    echo "Seed user does not exist. Registering..."

    request \
        POST \
        "$BASE_URL/register" \
        "" \
        "{
            \"email\": \"$SEED_EMAIL\",
            \"name\": \"Seed\",
            \"surname\": \"User\",
            \"phone\": \"+79990000000\",
            \"password\": \"$SEED_PASSWORD\"
        }"

    if [[ "$HTTP_STATUS" != "201" && "$HTTP_STATUS" != "409" ]]; then
        echo "$HTTP_BODY"
        fail "Cannot register seed user. HTTP $HTTP_STATUS"
    fi

    # В любом случае делаем login ещё раз.
    request \
        POST \
        "$BASE_URL/auth" \
        "" \
        "{
            \"email\": \"$SEED_EMAIL\",
            \"password\": \"$SEED_PASSWORD\"
        }"

    expect_status 200

    TOKEN=$(echo "$HTTP_BODY" | jq -r '.access_token')

    [[ "$TOKEN" != "null" && -n "$TOKEN" ]] \
        || fail "Login response does not contain access_token"

    echo "Authentication successful"
}

# ============================================================
# Suppliers
# ============================================================

get_supplier_id() {
    local name="$1"

    request GET "$BASE_URL/suppliers" "$TOKEN"
    expect_status 200

    echo "$HTTP_BODY" |
        jq -r --arg name "$name" \
            '.[] | select(.name == $name) | .id' |
        head -n 1
}

create_supplier() {
    local name="$1"
    local phone="$2"
    local country="$3"
    local city="$4"
    local street="$5"

    local id
    id=$(get_supplier_id "$name")

    if [[ -n "$id" ]]; then
        echo "Supplier '$name' already exists: $id" >&2
        echo "$id"
        return
    fi

    echo "Creating supplier: $name" >&2

    request \
        POST \
        "$BASE_URL/suppliers" \
        "$TOKEN" \
        "{
            \"name\": \"$name\",
            \"phone_number\": \"$phone\",
            \"address\": {
                \"country\": \"$country\",
                \"city\": \"$city\",
                \"street\": \"$street\"
            }
        }"

    expect_status 201

    id=$(get_supplier_id "$name")

    [[ -n "$id" ]] || fail "Cannot obtain ID for supplier '$name'"

    echo "$id"
}

seed_suppliers() {
    log "Creating suppliers"

    TECH_SUPPLIER_ID=$(
        create_supplier \
            "TechSupply" \
            "+74951112233" \
            "Russia" \
            "Moscow" \
            "Baumanskaya street 5"
    )

    EURO_SUPPLIER_ID=$(
        create_supplier \
            "EuroElectronics" \
            "+31205550101" \
            "Netherlands" \
            "Amsterdam" \
            "Damrak 10"
    )

    HOME_SUPPLIER_ID=$(
        create_supplier \
            "HomeDevices" \
            "+49305550202" \
            "Germany" \
            "Berlin" \
            "Alexanderplatz 3"
    )

    echo
    echo "Supplier IDs:"
    echo "TechSupply:      $TECH_SUPPLIER_ID"
    echo "EuroElectronics: $EURO_SUPPLIER_ID"
    echo "HomeDevices:     $HOME_SUPPLIER_ID"
}

# ============================================================
# Products
# ============================================================

get_product_id() {
    local name="$1"

    request GET "$BASE_URL/products" "$TOKEN"
    expect_status 200

    echo "$HTTP_BODY" |
        jq -r --arg name "$name" \
            '.[] | select(.name == $name) | .id' |
        head -n 1
}

create_product() {
    local name="$1"
    local category="$2"
    local price="$3"
    local stock="$4"
    local supplier_id="$5"

    local id
    id=$(get_product_id "$name")

    if [[ -n "$id" ]]; then
        echo "Product '$name' already exists: $id" >&2
        echo "$id"
        return
    fi

    echo "Creating product: $name" >&2

    request \
        POST \
        "$BASE_URL/products" \
        "$TOKEN" \
        "{
            \"name\": \"$name\",
            \"category\": \"$category\",
            \"price\": $price,
            \"available_stock\": $stock,
            \"supplier_id\": \"$supplier_id\"
        }"

    expect_status 201

    id=$(get_product_id "$name")

    [[ -n "$id" ]] || fail "Cannot obtain ID for product '$name'"

    echo "$id"
}

seed_products() {
    log "Creating products"

    MACBOOK_ID=$(
        create_product \
            "MacBook Air M4" \
            "laptop" \
            109990 \
            8 \
            "$TECH_SUPPLIER_ID"
    )

    MONITOR_ID=$(
        create_product \
            "Samsung 4K Monitor" \
            "monitor" \
            42990 \
            14 \
            "$EURO_SUPPLIER_ID"
    )

    MOUSE_ID=$(
        create_product \
            "Logitech MX Master" \
            "mouse" \
            10990 \
            25 \
            "$HOME_SUPPLIER_ID"
    )

    KEYBOARD_ID=$(
        create_product \
            "Keychron K8 Pro" \
            "keyboard" \
            13990 \
            17 \
            "$TECH_SUPPLIER_ID"
    )

    HEADPHONES_ID=$(
        create_product \
            "Sony WH-1000XM5" \
            "headphones" \
            37990 \
            11 \
            "$EURO_SUPPLIER_ID"
    )

    WASHER_ID=$(
        create_product \
            "Bosch Serie 6" \
            "washing_machine" \
            69990 \
            5 \
            "$HOME_SUPPLIER_ID"
    )

    TV_ID=$(
        create_product \
            "LG Smart TV 55" \
            "tv" \
            74990 \
            9 \
            "$EURO_SUPPLIER_ID"
    )

    VACUUM_ID=$(
        create_product \
            "Dyson V15" \
            "vacuum_cleaner" \
            65990 \
            6 \
            "$TECH_SUPPLIER_ID"
    )

    echo
    echo "Product IDs:"
    echo "MacBook:     $MACBOOK_ID"
    echo "Monitor:     $MONITOR_ID"
    echo "Mouse:       $MOUSE_ID"
    echo "Keyboard:    $KEYBOARD_ID"
    echo "Headphones:  $HEADPHONES_ID"
    echo "Washer:      $WASHER_ID"
    echo "TV:          $TV_ID"
    echo "Vacuum:      $VACUUM_ID"
}

# ============================================================
# Clients
# ============================================================

client_exists() {
    local name="$1"
    local surname="$2"

    request \
        GET \
        "$BASE_URL/clients/search?name=$name&surname=$surname" \
        "$TOKEN"

    expect_status 200

    local count
    count=$(echo "$HTTP_BODY" | jq 'length')

    [[ "$count" -gt 0 ]]
}

create_client() {
    local name="$1"
    local surname="$2"
    local birthday="$3"
    local gender="$4"
    local country="$5"
    local city="$6"
    local street="$7"

    if client_exists "$name" "$surname"; then
        echo "Client '$name $surname' already exists"
        return
    fi

    echo "Creating client: $name $surname"

    request \
        POST \
        "$BASE_URL/clients" \
        "$TOKEN" \
        "{
            \"name\": \"$name\",
            \"surname\": \"$surname\",
            \"birthday\": \"$birthday\",
            \"gender\": \"$gender\",
            \"address\": {
                \"country\": \"$country\",
                \"city\": \"$city\",
                \"street\": \"$street\"
            }
        }"

    expect_status 201
}

seed_clients() {
    log "Creating clients"

    create_client \
        "Alexey" \
        "Ivanov" \
        "1992-03-14T00:00:00Z" \
        "male" \
        "Russia" \
        "Saint Petersburg" \
        "Nevsky prospect 28"

    create_client \
        "Maria" \
        "Petrova" \
        "1995-07-21T00:00:00Z" \
        "female" \
        "France" \
        "Paris" \
        "Rue de Rivoli 40"

    create_client \
        "John" \
        "Smith" \
        "1988-11-05T00:00:00Z" \
        "male" \
        "Netherlands" \
        "Amsterdam" \
        "Prinsengracht 100"

    create_client \
        "Anna" \
        "Sidorova" \
        "2000-01-18T00:00:00Z" \
        "female" \
        "Russia" \
        "Kazan" \
        "Baumana street 12"
}
# ============================================================
# Images
# ============================================================

product_has_image() {
    local product_id="$1"

    request GET "$BASE_URL/products/$product_id" "$TOKEN"
    expect_status 200

    local image_id
    image_id=$(echo "$HTTP_BODY" | jq -r '.image_id')

    [[ "$image_id" != "null" && -n "$image_id" ]]
}

add_image() {
    local product_id="$1"
    local name="$2"
    local image_file="$3"

    if product_has_image "$product_id"; then
        echo "Product '$name' already has an image"
        return
    fi

    echo "Adding image '$image_file' to: $name"

    local image_base64
    image_base64=$(base64 < "$image_file" | tr -d '\n')

    request \
        POST \
        "$BASE_URL/products/$product_id/image" \
        "$TOKEN" \
        "{
            \"image\": \"$image_base64\"
        }"

    expect_status 201
}

seed_images() {
    log "Adding product images"

    add_image \
        "$MACBOOK_ID" \
        "MacBook Air M4" \
        "${IMAGE_FILES[0]}"

    add_image \
        "$MONITOR_ID" \
        "Samsung 4K Monitor" \
        "${IMAGE_FILES[1]}"

    add_image \
        "$MOUSE_ID" \
        "Logitech MX Master" \
        "${IMAGE_FILES[2]}"

    add_image \
        "$KEYBOARD_ID" \
        "Keychron K8 Pro" \
        "${IMAGE_FILES[3]}"

    add_image \
        "$HEADPHONES_ID" \
        "Sony WH-1000XM5" \
        "${IMAGE_FILES[4]}"

    # Картинок 5, товаров 8 — используем их повторно по кругу.
    add_image \
        "$WASHER_ID" \
        "Bosch Serie 6" \
        "${IMAGE_FILES[0]}"

    add_image \
        "$TV_ID" \
        "LG Smart TV 55" \
        "${IMAGE_FILES[1]}"

    add_image \
        "$VACUUM_ID" \
        "Dyson V15" \
        "${IMAGE_FILES[2]}"
}


# ============================================================
# Main
# ============================================================

main() {
    check_dependencies

    echo "========================================"
    echo " Appliance Shop Database Seeder"
    echo "========================================"
    echo "API:   $BASE_URL"
    echo "Images:"
    for image_file in "${IMAGE_FILES[@]}"; do
        echo "  $image_file"
    done

    get_token
    seed_suppliers
    seed_products
    seed_clients
    seed_images
    verify_seed
}

main