CREATE TABLE `alert_contacts` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `realm_name` varchar(50) DEFAULT 'master',
  `name` varchar(255) NOT NULL,
  `type` varchar(50) NOT NULL,
  `enabled` tinyint(1) DEFAULT '1',
  `details` json DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_realm_name` (`realm_name`),
  KEY `idx_deleted_at` (`deleted_at`),
  CONSTRAINT `fk_alert_contacts_realms` FOREIGN KEY (`realm_name`) REFERENCES `realms` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;