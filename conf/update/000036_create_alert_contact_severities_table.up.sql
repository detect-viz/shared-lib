CREATE TABLE `alert_contact_severities` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `contact_id` bigint NOT NULL,
  `severity_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_contact_severity` (`contact_id`,`severity_id`),
  KEY `severity_id` (`severity_id`),
  CONSTRAINT `alert_contact_severities_ibfk_1` FOREIGN KEY (`contact_id`) REFERENCES `alert_contacts` (`id`) ON DELETE CASCADE,
  CONSTRAINT `alert_contact_severities_ibfk_2` FOREIGN KEY (`severity_id`) REFERENCES `alert_severities` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;