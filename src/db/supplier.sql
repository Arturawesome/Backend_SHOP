CREATE TABLE IF NOT EXISTS supplier (
    id UUID PRIMARY KEY,
	name  VARCHAR(100) NOT NULL,
    address_id UUID  NOT NULL  REFERENCES address(id),
    phone_number  VARCHAR(100) NOT NULL
);

