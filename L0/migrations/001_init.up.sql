CREATE TABLE orders (
    order_uid UUID PRIMARY KEY,
    track_number VARCHAR(255),
    entry VARCHAR(12),
    locale VARCHAR(6),
    internal_signature TEXT,
    customer_id VARCHAR(255),
    delivery_service VARCHAR(255),
    shardkey VARCHAR(12),
    sm_id BIGINT,
    date_created TIMESTAMP WITH TIME ZONE,
    oof_shard VARCHAR(12)
);

CREATE TABLE deliveries (
    order_uid UUID PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    name VARCHAR(255),
    phone VARCHAR(30),
    zip VARCHAR(20),
    city VARCHAR(100),
    address TEXT,
    region VARCHAR(100),
    email VARCHAR(255)
);

CREATE TABLE payments (
    order_uid UUID PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction VARCHAR(255),
    request_id VARCHAR(255),
    currency VARCHAR(3),
    provider VARCHAR(255),
    amount DECIMAL(12, 2),
    payment_dt BIGINT,
    bank VARCHAR(255),
    delivery_cost DECIMAL(12,2),
    goods_total DECIMAL(12, 2),
    custom_fee DECIMAL(12, 2)
);

CREATE TABLE items (
    chrt_id BIGINT PRIMARY KEY,
    order_uid UUID NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    track_number VARCHAR(255),
    price DECIMAL(12,2),
    rid VARCHAR(255),
    name TEXT,
    sale SMALLINT,
    size VARCHAR(50),
    total_price DECIMAL(12,2),
    nm_id BIGINT,
    brand VARCHAR(255),
    status SMALLINT
);


CREATE INDEX idx_items_order_uid ON items(order_uid);
