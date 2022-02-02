ALTER TABLE product_price ADD INDEX product_price_product_price_type (product, price_type) TYPE bloom_filter(0.01) GRANULARITY 2;OPTIMIZE TABLE product_price FINAL;
