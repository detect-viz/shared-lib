CREATE TABLE `alert_rule_labels` (
  `realm_name` varchar(20) NOT NULL DEFAULT 'master',
  `rule_id` bigint NOT NULL,
  `label_id` bigint NOT NULL,
  `created_at` bigint DEFAULT NULL,
  `updated_at` bigint DEFAULT NULL,
  FOREIGN KEY (`rule_id`) REFERENCES `alert_rules` (`id`),
  FOREIGN KEY (`label_id`) REFERENCES `labels` (`id`),
  FOREIGN KEY (`realm_name`) REFERENCES `realms` (`name`),
  UNIQUE KEY `idx_alert_rule_label` (`rule_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;