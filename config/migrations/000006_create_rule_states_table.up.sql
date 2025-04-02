CREATE TABLE `rule_states` (
  `rule_id` binary(16) NOT NULL,
  `silence_start_at` bigint DEFAULT NULL,
  `silence_end_at` bigint DEFAULT NULL,
  `last_triggered_severity` enum('info','warn','crit') DEFAULT NULL,
  `state` enum('alerting','resolved','normal','disabled') DEFAULT 'normal',
  `contact_state` enum('normal','muting','silence','delayed') DEFAULT 'normal',
  `contact_counter` int DEFAULT '0',
  `first_triggered_at` bigint DEFAULT NULL,
  `last_triggered_at` bigint DEFAULT NULL,
  `last_check_value` decimal(10,2) DEFAULT NULL,
  `last_triggered_value` decimal(10,2) DEFAULT NULL,
  `last_triggered_log_id` binary(16) DEFAULT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `updated_at` bigint unsigned DEFAULT NULL,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`rule_id`),
  FOREIGN KEY (`rule_id`) REFERENCES `rules` (`id`) ON DELETE CASCADE,
  KEY `idx_silence_period` (`silence_start_at`, `silence_end_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


