CREATE TABLE `notify_trigger_logs` (
  `notify_log_uuid` VARCHAR(36) NOT NULL,
  `trigger_log_uuid` VARCHAR(36) NOT NULL,
  PRIMARY KEY (`notify_log_uuid`,`trigger_log_uuid`),
  KEY `trigger_log_uuid` (`trigger_log_uuid`),
  CONSTRAINT `notify_trigger_logs_ibfk_1` FOREIGN KEY (`notify_log_uuid`) REFERENCES `notify_logs` (`uuid`) ON DELETE CASCADE,
  CONSTRAINT `notify_trigger_logs_ibfk_2` FOREIGN KEY (`trigger_log_uuid`) REFERENCES `trigger_logs` (`uuid`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;