CREATE TABLE IF NOT EXISTS versions (
    id SERIAL PRIMARY KEY,
    config_id INT REFERENCES configs(id) ON DELETE CASCADE,
    parent_vid INT REFERENCES versions(id) ON DELETE SET NULL,
    config_val JSONB NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    author VARCHAR(255) NOT NULL,
    message TEXT
);

CREATE TABLE IF NOT EXISTS active_pointers (
    config_id INT PRIMARY KEY REFERENCES configs(id) ON DELETE CASCADE,
    active_vid INT REFERENCES versions(id) ON DELETE SET NULL,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);