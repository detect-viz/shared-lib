CREATE TABLE `alert_rule_details` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `realm_name` varchar(20) NOT NULL DEFAULT 'master',
  `alert_rule_id` bigint NOT NULL,
  `resource_name` varchar(100) NOT NULL,
  `partition_name` varchar(50) DEFAULT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `updated_at` bigint unsigned DEFAULT NULL,
  `deleted_at` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_detail_realms` (`realm_name`),
  KEY `fk_detail_rules` (`alert_rule_id`),
  CONSTRAINT `fk_detail_realms` FOREIGN KEY (`realm_name`) REFERENCES `realms` (`name`),
  CONSTRAINT `fk_detail_rules` FOREIGN KEY (`alert_rule_id`) REFERENCES `alert_rules` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;