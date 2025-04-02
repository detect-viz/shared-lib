CREATE TABLE `mute_resource_groups` (
  `mute_id` VARCHAR(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `resource_group_id` BIGINT NOT NULL,
  PRIMARY KEY (`mute_id`, `resource_group_id`),
  KEY `fk_mute_resource_groups_group` (`resource_group_id`),
  CONSTRAINT `fk_mute_resource_groups_group` FOREIGN KEY (`resource_group_id`) REFERENCES `resource_groups` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_mute_resource_groups_mute` FOREIGN KEY (`mute_id`) REFERENCES `mutes` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB;
