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