CREATE TABLE `rule_contacts` (
  `rule_id` VARCHAR(36) NOT NULL,
  `contact_id` bigint NOT NULL,
  UNIQUE KEY `uk_rule_contact` (`rule_id`,`contact_id`),
  KEY `fk_rule_contact` (`rule_id`),
  KEY `fk_contact_rule` (`contact_id`),
  CONSTRAINT `fk_contact_rule` FOREIGN KEY (`contact_id`) REFERENCES `contacts` (`id`),
  CONSTRAINT `fk_rule_contact` FOREIGN KEY (`rule_id`) REFERENCES `rules` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB;