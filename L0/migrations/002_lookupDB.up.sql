CREATE TABLE order_shard_maping(
    order_uid UUID PRIMARY KEY,
    shardkey VARCHAR(12)
);