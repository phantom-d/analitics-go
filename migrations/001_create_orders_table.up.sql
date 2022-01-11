CREATE TABLE IF NOT EXISTS product_price (
	product UUID,
	price UInt32 DEFAULT 0,
	price_type UUID,
	price_date Date,
	price_time DateTime('Europe/Moscow')
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(price_date)
ORDER BY (price_date, price_time);
