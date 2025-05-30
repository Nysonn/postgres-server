CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    price_ugx DECIMAL(12,2) NOT NULL,
    available BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add some sample data
INSERT INTO items (name, category, price_ugx, available) VALUES
    ('Test Item 1', 'Electronics', 100000.00, true),
    ('Test Item 2', 'Clothing', 50000.00, true),
    ('Test Item 3', 'Books', 25000.00, false); 