CREATE TABLE `rule_details` (
  `id` VARCHAR(36) NOT NULL,
  `rule_id` VARCHAR(36) NOT NULL,
  `resource_name` varchar(100) NOT NULL,
  `partition_name` varchar(50) DEFAULT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `updated_at` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_rule_id_resource_name_partition_name` (`rule_id`, `resource_name`, `partition_name`),
  KEY `fk_detail_rules` (`rule_id`),
  CONSTRAINT `fk_detail_rules` FOREIGN KEY (`rule_id`) REFERENCES `rules` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB;