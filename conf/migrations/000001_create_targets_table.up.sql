CREATE TABLE `targets` (
  `realm_name` varchar(20) NOT NULL,
  `id` binary(16) NOT NULL,
  `is_hidden` tinyint(1) NOT NULL DEFAULT '0',
  `datasource_type` varchar(50) NOT NULL,
  `category` varchar(50) NOT NULL,
  `resource_name` varchar(100) NOT NULL,
  `partition_name` varchar(255) DEFAULT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `updated_at` bigint unsigned DEFAULT NULL,
  `deleted_at` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_targets_realm` (`realm_name`),
  CONSTRAINT `fk_realms_targets` FOREIGN KEY (`realm_name`) REFERENCES `realms` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;