CREATE TABLE `parser_awrrpt_files` (
  `realm_name` varchar(20) NOT NULL,
  `host` varchar(50) NOT NULL,
  `db_name` varchar(50) NOT NULL,
  `date` date NOT NULL,
  `hour` tinyint NOT NULL,
  `content` mediumblob NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`realm_name`,`host`,`db_name`,`date`,`hour`),
  KEY `idx_awrrpt_files_date` (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `parser_awrrpt_sqls` (
  `realm_name` varchar(20) NOT NULL,
  `db_name` varchar(50) NOT NULL,
  `id` varchar(255) NOT NULL,
  `text` longtext NOT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_awrrpt_sqls_realm` (`realm_name`),
  CONSTRAINT `fk_sqls_realms` FOREIGN KEY (`realm_name`) REFERENCES `realms` (`name`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
