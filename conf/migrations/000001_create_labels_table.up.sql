CREATE TABLE labels (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    realm_name VARCHAR(255) NOT NULL,
    key_name VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    FOREIGN KEY (realm_name) REFERENCES realms(name),
    UNIQUE KEY uk_label (realm_name, key_name),
    KEY idx_label_key (key_name)
);
