CREATE TABLE `contacts` (
  `realm_name` varchar(20) NOT NULL,
  `id` binary(16) NOT NULL,
  `channel_type` varchar(50) NOT NULL,
  `name` varchar(100) NOT NULL,
  `enabled` tinyint(1) DEFAULT '1',
  `send_resolved` tinyint(1) DEFAULT '1',
  `max_retry` int DEFAULT '3',
  `retry_delay` varchar(10) NOT NULL DEFAULT '5m',
  `severities` set('info', 'warn', 'crit') NOT NULL,
  `config` json DEFAULT NULL,
  `created_by` varchar(36) DEFAULT NULL,
  `updated_by` varchar(36) DEFAULT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `updated_at` bigint unsigned DEFAULT NULL,
  `deleted_at` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_contacts_realm` (`realm_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;