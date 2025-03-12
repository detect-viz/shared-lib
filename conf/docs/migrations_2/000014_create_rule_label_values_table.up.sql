CREATE TABLE `rule_label_values` (
  `rule_id` VARCHAR(36) NOT NULL,
  `label_value_id` bigint NOT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `updated_at` bigint unsigned DEFAULT NULL,
  FOREIGN KEY (`rule_id`) REFERENCES `rules` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  FOREIGN KEY (`label_value_id`) REFERENCES `label_values` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  UNIQUE KEY `idx_rule_label` (`rule_id`, `label_value_id`),
  INDEX `idx_rule_label_values` (`rule_id`, `label_value_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;