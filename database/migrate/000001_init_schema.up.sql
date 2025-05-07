CREATE TABLE IF NOT EXISTS urls (
    "id" BIGINT AUTO_INCREMENT PRIMARY KEY,
    "original_url" TEXT NOT NULL,
    "short_code" TEXT NOT NULL UNIQUE,
    "is_custom" BOOLEAN NOT NULL DEFAULT FALSE,
    "expired_at" TIMESTAMP NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_short_code ON urls(short_code);
CREATE INDEX idx_expired_at ON urls(expired_at);