CREATE TABLE `notification_trigger_logs` (
  `notification_id` bigint NOT NULL,
  `trigger_log_id` bigint NOT NULL,
  PRIMARY KEY (`notification_id`,`trigger_log_id`),
  KEY `trigger_log_id` (`trigger_log_id`),
  CONSTRAINT `notification_trigger_logs_ibfk_1` FOREIGN KEY (`notification_id`) REFERENCES `notification_logs` (`id`) ON DELETE CASCADE,
  CONSTRAINT `notification_trigger_logs_ibfk_2` FOREIGN KEY (`trigger_log_id`) REFERENCES `trigger_logs` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;