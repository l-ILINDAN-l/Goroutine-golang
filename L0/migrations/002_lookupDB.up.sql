CREATE TABLE order_shard_mapping(
    order_uid VARCHAR(19) PRIMARY KEY,
    shard_key VARCHAR(12)
);