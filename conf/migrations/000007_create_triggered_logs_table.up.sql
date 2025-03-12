CREATE TABLE `triggered_logs` (
  `realm_name` varchar(20) NOT NULL,
  `id` binary(16) NOT NULL,
  `triggered_at` bigint NOT NULL,
  `last_triggered_at` bigint NOT NULL,
  `resolved_at` bigint DEFAULT NULL,
  `notify_log_id` binary(16) DEFAULT NULL,
  `resolved_notify_log_id` binary(16) DEFAULT NULL,
  `notify_state` varchar(20) DEFAULT NULL,
  `resolved_notify_state` varchar(20) DEFAULT NULL,
  `resource_name` varchar(100) NOT NULL,
  `partition_name` varchar(255) DEFAULT NULL,
  `metric_rule_uid` varchar(100) NOT NULL,
  `rule_id` binary(16) NOT NULL,
  `rule_snapshot` json DEFAULT NULL,
  `rule_state_snapshot` json DEFAULT NULL,
  `severity` enum('info','warn','crit') NOT NULL,
  `triggered_value` decimal(10,2) NOT NULL,
  `resolved_value` decimal(10,2) DEFAULT NULL,
  `threshold` decimal(10,2) NOT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `updated_at` bigint unsigned DEFAULT NULL,
  `deleted_at` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_triggered_at` (`triggered_at`),
  KEY `idx_rule_id` (`rule_id`),
  KEY `idx_notify_state` (`notify_state`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
