CREATE TABLE `alert_contact_severities` (
  `alert_contact_id` bigint NOT NULL,
  `severity` enum('info','warn','crit') NOT NULL,
  UNIQUE KEY `idx_contact_severity` (`alert_contact_id`,`severity`),
  CONSTRAINT `alert_contact_severities_ibfk_1` FOREIGN KEY (`alert_contact_id`) REFERENCES `alert_contacts` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;