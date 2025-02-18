CREATE TABLE `alert_states` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `rule_id` bigint NOT NULL,
  `resource_name` varchar(50) NOT NULL,
  `metric_name` varchar(50) NOT NULL,
  `start_time` bigint NOT NULL,
  `last_time` bigint NOT NULL,
  `last_value` decimal(10,2) DEFAULT '0.00',
  `stack_duration` int DEFAULT '0',
  `previous_value` decimal(10,2) DEFAULT '0.00',
  `amplitude` decimal(10,2) DEFAULT '0.00',
  `created_at` bigint DEFAULT NULL,
  `updated_at` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_alert_state` (`rule_id`,`resource_name`,`metric_name`),
  KEY `idx_start_time` (`start_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;