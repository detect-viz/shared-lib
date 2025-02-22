CREATE TABLE `alert_templates` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `realm_name` varchar(50) DEFAULT 'master',
  `is_default` tinyint(1) DEFAULT '0',
  `name` varchar(255) NOT NULL,
  `format` enum('html','text','markdown','json') NOT NULL,
  `rule_state` enum('alerting','resolved','normal','disabled') NOT NULL,
  `title_template` text NOT NULL,
  `message_template` text NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_default` (`is_default`),
  KEY `idx_realm_name` (`realm_name`),
  CONSTRAINT `fk_alert_templates_realms` FOREIGN KEY (`realm_name`) REFERENCES `realms` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;