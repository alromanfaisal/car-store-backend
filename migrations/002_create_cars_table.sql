-- migrations/002_create_cars_table.sql
CREATE TABLE IF NOT EXISTS cars (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    image_url TEXT NOT NULL,
    price INTEGER NOT NULL,
    discount_price INTEGER,
    is_new BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);