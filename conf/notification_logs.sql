CREATE TABLE `trigger_logs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `uuid` varchar(36) NOT NULL,
  `timestamp` bigint NOT NULL,
  `rule_id` varchar(50) NOT NULL,
  `rule_name` varchar(255) NOT NULL,
  `resource_group` varchar(100) NOT NULL,
  `resource_name` varchar(100) NOT NULL,
  `partition_name` varchar(100) DEFAULT NULL,
  `metric` varchar(100) NOT NULL,
  `value` decimal(10,2) NOT NULL,
  `threshold` decimal(10,2) NOT NULL,
  `unit` varchar(10) NOT NULL,
  `severity` enum('info','warn','crit') NOT NULL,
  `duration` int NOT NULL,
  `status` enum('alerting','resolved') NOT NULL DEFAULT 'alerting',
  `notify_status` enum('pending','sent','failed') NOT NULL DEFAULT 'pending',
  `notify_retry` int DEFAULT '0',
  `silence_start` bigint DEFAULT NULL,
  `silence_end` bigint DEFAULT NULL,
  `resolved_time` bigint DEFAULT NULL,
  `labels` json DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_timestamp` (`timestamp`),
  KEY `idx_rule_id` (`rule_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `notification_logs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `realm_name` varchar(50) NOT NULL,
  `contact_id` bigint NOT NULL,
  `contact_name` varchar(100) NOT NULL,
  `contact_type` varchar(50) NOT NULL,
  `subject` text NOT NULL,
  `file_path` varchar(255) NOT NULL,
  `sent_at` bigint NOT NULL,
  `status` enum('sent','failed','pending') NOT NULL DEFAULT 'pending',
  `error` text,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `fk_notification_logs_realms` (`realm_name`),
  CONSTRAINT `fk_notification_logs_realms` FOREIGN KEY (`realm_name`) REFERENCES `realms` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `notification_trigger_logs` (
  `notification_id` bigint NOT NULL,
  `trigger_log_id` bigint NOT NULL,
  PRIMARY KEY (`notification_id`,`trigger_log_id`),
  KEY `trigger_log_id` (`trigger_log_id`),
  CONSTRAINT `notification_trigger_logs_ibfk_1` FOREIGN KEY (`notification_id`) REFERENCES `notification_logs` (`id`) ON DELETE CASCADE,
  CONSTRAINT `notification_trigger_logs_ibfk_2` FOREIGN KEY (`trigger_log_id`) REFERENCES `trigger_logs` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


CREATE TABLE `alert_state` (
  `rule_id` bigint NOT NULL AUTO_INCREMENT,
  `resource_name` varchar(50) NOT NULL,
  `metric_name` varchar(50) NOT NULL,
  `start_time` bigint NOT NULL,
  `last_time` bigint NOT NULL,
  `last_value` decimal(10,2) DEFAULT '0.00',
  `stack_duration` int DEFAULT '0',
  `previous_value` decimal(10,2) DEFAULT '0.00',
  `amplitude` decimal(10,2) DEFAULT '0.00',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`rule_id`),
  UNIQUE KEY `idx_alert_state` (`resource_name`,`metric_name`),
  KEY `idx_start_time` (`start_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;