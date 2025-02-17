CREATE TABLE `alert_severities` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `severity` enum('level_info','level_warn','level_crit') NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `severity` (`severity`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;