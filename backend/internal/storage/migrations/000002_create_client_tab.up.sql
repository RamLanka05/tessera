CREATE TABLE IF NOT EXISTS clients (
    id SERIAL PRIMARY KEY,
    client_id VARCHAR(255) UNIQUE NOT NULL,
    client_secret VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO clients (client_id, client_secret) VALUES
    ('client1', 'secret1'),
    ('client2', 'secret2')
ON CONFLICT (client_id) DO NOTHING;