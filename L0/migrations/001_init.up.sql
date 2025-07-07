CREATE TABLE orders (
    order_uid VARCHAR(19) PRIMARY KEY, -- Раз в model.json 19 символов, а не 36
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
    order_uid VARCHAR(19) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    name VARCHAR(255),
    phone VARCHAR(30),
    zip VARCHAR(20),
    city VARCHAR(100),
    address TEXT,
    region VARCHAR(100),
    email VARCHAR(255)
);

CREATE TABLE payments (
    order_uid VARCHAR(19) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction VARCHAR(255),
    request_id VARCHAR(255),
    currency VARCHAR(3),
    provider VARCHAR(255),
    amount BIGINT,
    payment_dt BIGINT,
    bank VARCHAR(255),
    delivery_cost BIGINT,
    goods_total BIGINT,
    custom_fee BIGINT
);

CREATE TABLE items (
    chrt_id BIGINT PRIMARY KEY,
    order_uid VARCHAR(19) NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    track_number VARCHAR(255),
    price BIGINT,
    rid VARCHAR(255),
    name TEXT,
    sale SMALLINT,
    size VARCHAR(50),
    total_price BIGINT,
    nm_id BIGINT,
    brand VARCHAR(255),
    status SMALLINT
);


CREATE INDEX idx_items_order_uid ON items(order_uid);
