CREATE TABLE `notify_logs` (
  `realm_name` varchar(20) NOT NULL,
  `id` binary(16) NOT NULL,
  `sent_at` bigint DEFAULT NULL,
  `state` varchar(20) NOT NULL DEFAULT 'pending',
  `retry_counter` int DEFAULT '0',
  `last_retry_at` bigint DEFAULT NULL,
  `error_messages` json DEFAULT NULL,
  `triggered_log_ids` json DEFAULT NULL,
  `contact_id` binary(16) NOT NULL,
  `channel_type` varchar(50) NOT NULL,
  `contact_snapshot` json DEFAULT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `updated_at` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_notify_logs_contact` (`contact_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;