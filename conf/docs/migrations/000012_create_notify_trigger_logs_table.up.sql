CREATE TABLE `notify_triggered_logs` (
  `notify_log_id` VARCHAR(36) NOT NULL,
  `triggered_log_id` VARCHAR(36) NOT NULL,
  PRIMARY KEY (`notify_log_id`,`triggered_log_id`),
  KEY `triggered_log_id` (`triggered_log_id`),
  CONSTRAINT `notify_triggered_logs_ibfk_1` FOREIGN KEY (`notify_log_id`) REFERENCES `notify_logs` (`id`) ON DELETE CASCADE,
  CONSTRAINT `notify_triggered_logs_ibfk_2` FOREIGN KEY (`triggered_log_id`) REFERENCES `triggered_logs` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;