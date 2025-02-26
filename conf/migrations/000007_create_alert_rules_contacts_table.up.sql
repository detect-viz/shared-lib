CREATE TABLE `alert_rule_contacts` (
  `alert_rule_id` bigint NOT NULL,
  `alert_contact_id` bigint NOT NULL,
  UNIQUE KEY `uk_rule_contact` (`alert_rule_id`,`alert_contact_id`),
  KEY `fk_rule_contact` (`alert_rule_id`),
  KEY `fk_contact_rule` (`alert_contact_id`),
  CONSTRAINT `fk_contact_rule` FOREIGN KEY (`alert_contact_id`) REFERENCES `contacts` (`id`),
  CONSTRAINT `fk_rule_contact` FOREIGN KEY (`alert_rule_id`) REFERENCES `alert_rules` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;