CREATE TABLE `templates` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `realm_name` varchar(20) DEFAULT 'master',
  `name` varchar(100) NOT NULL,
  `title` text NOT NULL,
  `message` text NOT NULL,
  `create_type` enum('system','user') NOT NULL,
  `format_type` enum('html','text','markdown','json') NOT NULL,
  `rule_state` enum('alerting','resolved') NOT NULL,
  `created_by` varchar(36) DEFAULT NULL,
  `updated_by` varchar(36) DEFAULT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `updated_at` bigint unsigned DEFAULT NULL,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_realm_name_name` (`realm_name`, `name`),
  KEY `idx_realm_name` (`realm_name`),
  CONSTRAINT `fk_templates_realms` FOREIGN KEY (`realm_name`) REFERENCES `realms` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;