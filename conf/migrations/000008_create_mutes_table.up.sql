CREATE TABLE `mutes` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `realm_name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'master',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '抑制規則名稱',
  `years` json DEFAULT NULL COMMENT '指定年份範圍 (如 ["2020:2022", "2030"])',
  `time_intervals` json NOT NULL COMMENT '一天內多個時間區間 (如 [{"start_time": "06:00", "end_time": "23:59"}])',
  `repeat_type` enum('never','daily','weekly','monthly') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'never' COMMENT '重複類型',
  `weekdays` json DEFAULT NULL COMMENT '允許的星期 (如 ["monday:wednesday", "saturday"])',
  `months` json DEFAULT NULL COMMENT '允許的月份 (如 ["1:3", "may:august", "december"])',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '建立時間',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新時間',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '刪除時間 (軟刪除)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='告警抑制規則';