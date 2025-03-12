CREATE TABLE `contact_severities` (
  `contact_id` bigint NOT NULL,
  `severity` enum('info','warn','crit') NOT NULL,
  UNIQUE KEY `idx_contact_severity` (`contact_id`,`severity`),
  CONSTRAINT `contact_severities_ibfk_1` FOREIGN KEY (`contact_id`) REFERENCES `contacts` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;