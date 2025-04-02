



CREATE TABLE `realms` (
  `name` varchar(20) NOT NULL DEFAULT 'master',
  PRIMARY KEY (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `targets` (
  `realm_name` varchar(20) NOT NULL,
  `id` binary(16) NOT NULL,
  `is_hidden` tinyint(1) NOT NULL DEFAULT '0',
  `datasource_name` varchar(50) NOT NULL,
  `category` varchar(50) NOT NULL,
  `resource_name` varchar(100) NOT NULL,
  `partition_name` varchar(255) DEFAULT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `updated_at` bigint unsigned DEFAULT NULL,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_targets_realm` (`realm_name`),
  CONSTRAINT `fk_realms_targets` FOREIGN KEY (`realm_name`) REFERENCES `realms` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `contacts` (
  `realm_name` varchar(20) NOT NULL,
  `id` binary(16) NOT NULL,
  `channel_type` varchar(50) NOT NULL,
  `name` varchar(100) NOT NULL,
  `enabled` tinyint(1) DEFAULT '1',
  `send_resolved` tinyint(1) DEFAULT '1',
  `max_retry` int DEFAULT '3',
  `retry_delay` varchar(10) NOT NULL DEFAULT '5m',
  `severities` set('info', 'warn', 'crit') NOT NULL,
  `config` json DEFAULT NULL,
  `created_by` varchar(36) DEFAULT NULL,
  `updated_by` varchar(36) DEFAULT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `updated_at` bigint unsigned DEFAULT NULL,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_contacts_realm` (`realm_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `templates` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `realm_name` varchar(20) DEFAULT 'master',
  `name` varchar(100) NOT NULL,
  `title` text NOT NULL,
  `message` text NOT NULL,
  `create_type` enum('system','user') NOT NULL,
  `format_type` enum('html','text','markdown','json') NOT NULL,
  `rule_state` enum('alerting','resolved') NOT NULL,
  `created_by` varchar(36) DEFAULT NULL,
  `updated_by` varchar(36) DEFAULT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `updated_at` bigint unsigned DEFAULT NULL,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_realm_name_name` (`realm_name`, `name`),
  KEY `idx_realm_name` (`realm_name`),
  CONSTRAINT `fk_templates_realms` FOREIGN KEY (`realm_name`) REFERENCES `realms` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `rules` (
  `realm_name` varchar(20) NOT NULL,
  `id` binary(16) NOT NULL,
  `create_type` enum('user','system') NOT NULL DEFAULT 'user',
  `enabled` tinyint(1) NOT NULL DEFAULT '1',
  `auto_apply` tinyint(1) NOT NULL DEFAULT '0',
  `target_id` binary(16) NOT NULL,
  `metric_rule_uid` varchar(50) NOT NULL,
  `info_threshold` double DEFAULT NULL,
  `warn_threshold` double DEFAULT NULL ,
  `crit_threshold` double DEFAULT NULL,
  `duration` varchar(10) NOT NULL DEFAULT '5m',
  `silence_period` varchar(10) NOT NULL DEFAULT '1h',
  `times` int NOT NULL DEFAULT '3',
  `created_by` varchar(36) DEFAULT NULL,
  `updated_by` varchar(36) DEFAULT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `updated_at` bigint unsigned DEFAULT NULL,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_rule` (`realm_name`,`metric_rule_uid`,`target_id`),
  KEY `fk_realms_rules` (`realm_name`),
  CONSTRAINT `fk_realms_rules` FOREIGN KEY (`realm_name`) REFERENCES `realms` (`name`),
  CONSTRAINT `fk_targets_rules` FOREIGN KEY (`target_id`) REFERENCES `targets` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `rule_contacts` (
  `rule_id` binary(16) NOT NULL,
  `contact_id` binary(16) NOT NULL,
  PRIMARY KEY (`rule_id`, `contact_id`),
  CONSTRAINT `fk_rule_contacts_rule` FOREIGN KEY (`rule_id`) REFERENCES `rules` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_rule_contacts_contact` FOREIGN KEY (`contact_id`) REFERENCES `contacts` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `rule_states` (
  `rule_id` binary(16) NOT NULL,
  `silence_start_at` bigint DEFAULT NULL,
  `silence_end_at` bigint DEFAULT NULL,
  `last_triggered_severity` enum('info','warn','crit') DEFAULT NULL,
  `rule_state` enum('alerting','resolved','normal','disabled') DEFAULT 'normal',
  `contact_state` enum('normal','muting','silence','delayed') DEFAULT 'normal',
  `contact_counter` int DEFAULT '0',
  `first_triggered_time` bigint DEFAULT NULL,
  `last_triggered_time` bigint DEFAULT NULL,
  `last_check_value` decimal(10,2) DEFAULT NULL,
  `last_triggered_value` decimal(10,2) DEFAULT NULL,
  `amplitude` decimal(10,2) DEFAULT '0.00',
  `created_at` bigint unsigned DEFAULT NULL,
  `updated_at` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`rule_id`),
  FOREIGN KEY (`rule_id`) REFERENCES `rules` (`id`) ON DELETE CASCADE,
  KEY `idx_silence_period` (`silence_start_at`, `silence_end_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `triggered_logs` (
  `realm_name` varchar(20) NOT NULL,
  `id` binary(16) NOT NULL,
  `triggered_at` bigint NOT NULL,
  `last_triggered_at` bigint NOT NULL,
  `resolved_at` bigint DEFAULT NULL,
  `notify_log_id` binary(16) DEFAULT NULL,
  `resolved_notify_log_id` binary(16) DEFAULT NULL,
  `notify_state` varchar(20) DEFAULT NULL,
  `resolved_notify_state` varchar(20) DEFAULT NULL,
  `resource_name` varchar(100) NOT NULL,
  `partition_name` varchar(255) DEFAULT NULL,
  `metric_rule_uid` varchar(100) NOT NULL,
  `rule_id` binary(16) NOT NULL,
  `rule_snapshot` json DEFAULT NULL,
  `rule_state_snapshot` json DEFAULT NULL,
  `severity` enum('info','warn','crit') NOT NULL,
  `triggered_value` decimal(10,2) NOT NULL,
  `resolved_value` decimal(10,2) DEFAULT NULL,
  `threshold` decimal(10,2) NOT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `updated_at` bigint unsigned DEFAULT NULL,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_triggered_at` (`triggered_at`),
  KEY `idx_rule_id` (`rule_id`),
  KEY `idx_notify_state` (`notify_state`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `notify_logs` (
  `realm_name` varchar(20) NOT NULL,
  `id` binary(16) NOT NULL,
  `sent_at` bigint DEFAULT NULL,
  `state` varchar(20) NOT NULL DEFAULT 'pending',
  `retry_counter` int DEFAULT '0',
  `last_retry_at` bigint DEFAULT NULL,
  `error_messages` json DEFAULT NULL,
  `triggered_log_ids` json DEFAULT NULL,
  `contact_id` binary(16) NOT NULL,
  `channel_type` varchar(50) NOT NULL,
  `contact_snapshot` json DEFAULT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `updated_at` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_notify_logs_contact` (`contact_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


