DROP TABLE IF EXISTS `realms`;

CREATE TABLE `realms` (
  `name`       varchar(20) NOT NULL,
  `created_at` bigint      unsigned DEFAULT NULL,
  `updated_at` bigint      unsigned DEFAULT NULL,
  `deleted_at` bigint      DEFAULT NULL,
  PRIMARY KEY (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `realms` (`name`)
VALUES ('master');

DROP TABLE IF EXISTS `resource_groups`;

CREATE TABLE `resource_groups` (
  `realm_name` varchar(20) DEFAULT NULL,
  `type` varchar(20) NOT NULL,
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `updated_at` bigint unsigned DEFAULT NULL,
  `deleted_at` bigint DEFAULT NULL,
  `realm_groups` json NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`type`,`name`,`realm_name`),
  KEY `id` (`realm_name`),
  CONSTRAINT `id` FOREIGN KEY (`realm_name`) REFERENCES `realms` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `resources`;


CREATE TABLE `resources` (
  `name` varchar(100) NOT NULL,
  `resource_group_id` bigint NOT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `updated_at` bigint unsigned DEFAULT NULL,
  `deleted_at` bigint DEFAULT NULL,
  UNIQUE KEY `name` (`name`,`resource_group_id`),
  KEY `fk_resource_groups_resources` (`resource_group_id`),
  CONSTRAINT `fk_resource_groups_resources` FOREIGN KEY (`resource_group_id`) REFERENCES `resource_groups` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;