CREATE TABLE `labels` (
  `realm_name` varchar(20) NOT NULL DEFAULT 'master',
  `id` bigint NOT NULL AUTO_INCREMENT,
  `key` varchar(100) NOT NULL,
  `value` varchar(100) NOT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `updated_at` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_label` (`realm_name`,`key`,`value`),
  KEY `fk_labels_realms` (`realm_name`),
  CONSTRAINT `fk_labels_realms` FOREIGN KEY (`realm_name`) REFERENCES `realms` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;