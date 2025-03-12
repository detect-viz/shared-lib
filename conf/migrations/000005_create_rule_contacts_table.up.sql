CREATE TABLE `rule_contacts` (
  `rule_id` binary(16) NOT NULL,
  `contact_id` binary(16) NOT NULL,
  PRIMARY KEY (`rule_id`, `contact_id`),
  CONSTRAINT `fk_rule_contacts_rule` FOREIGN KEY (`rule_id`) REFERENCES `rules` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_rule_contacts_contact` FOREIGN KEY (`contact_id`) REFERENCES `contacts` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;