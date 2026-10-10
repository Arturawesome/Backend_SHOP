CREATE TABLE IF NOT EXISTS client (
    id UUID PRIMARY KEY,
    client_name  VARCHAR(100) NOT NULL,
    client_surname  VARCHAR(100)  NOT NULL,
    birthday DATE NOT NULL,
    gender  VARCHAR(100) NOT NULL,
    registration_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    address_id UUID  NOT NULL REFERENCES address(id)
);
