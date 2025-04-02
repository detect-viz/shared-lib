CREATE TABLE `label_keys` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `realm_name` VARCHAR(20) NOT NULL DEFAULT 'master',
  `key_name` VARCHAR(255) NOT NULL UNIQUE,
  `created_at` bigint unsigned NOT NULL,
  `updated_at` bigint unsigned NOT NULL,
  FOREIGN KEY (`realm_name`) REFERENCES `realms` (`name`) ON DELETE CASCADE,
  INDEX `idx_realm_key` (`realm_name`, `key_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `label_values` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  `label_key_id` BIGINT NOT NULL,
  `value` VARCHAR(255) NOT NULL,
  INDEX `idx_label_key_id` (`label_key_id`),
  CONSTRAINT `fk_label_values_key` FOREIGN KEY (`label_key_id`) REFERENCES `label_keys` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
