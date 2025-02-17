CREATE TABLE `alert_mute_rules` (
  `alert_mute_id` bigint NOT NULL,
  `alert_rule_id` bigint NOT NULL,
  PRIMARY KEY (`alert_mute_id`,`alert_rule_id`),
  KEY `fk_mute_rules_rule` (`alert_rule_id`),
  CONSTRAINT `fk_mute_rules_mute` FOREIGN KEY (`alert_mute_id`) REFERENCES `alert_mutes` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_mute_rules_rule` FOREIGN KEY (`alert_rule_id`) REFERENCES `alert_rules` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;