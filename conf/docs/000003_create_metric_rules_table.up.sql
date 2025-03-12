CREATE TABLE `metric_rules` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(100) NOT NULL COMMENT '告警規則名稱',
  `category` VARCHAR(50) NOT NULL COMMENT '類別 (cpu/memory/disk/network/database/filesystem)',
  `check_type` ENUM('absolute','amplitude') NOT NULL COMMENT '檢查類型: 絕對值 / 變化幅度',
  `metric_name` VARCHAR(50) NOT NULL COMMENT '監控指標名稱',
  `tags` VARCHAR(255) NOT NULL COMMENT '分區標籤: host,cpu / host,filesystem / host,database,tablespace',
  `raw_unit` VARCHAR(20) NOT NULL COMMENT '原始單位 (bytes, %, MB, GB)',
  `display_unit` VARCHAR(20) NOT NULL COMMENT '顯示單位 (GB, MB/s, %)',
  `scale` DOUBLE NOT NULL DEFAULT 1.0 COMMENT '轉換倍率: raw_value × scale = display_value',
  `operator` ENUM('gt','ge','lt','le') NOT NULL COMMENT '比較運算符: greater_than, less_than',
  `info_threshold` DOUBLE DEFAULT NULL COMMENT '資訊等級閾值 (可選)',
  `warn_threshold` DOUBLE DEFAULT NULL COMMENT '警告等級閾值',
  `crit_threshold` DOUBLE DEFAULT NULL COMMENT '嚴重等級閾值',
  `duration` INT DEFAULT NULL COMMENT '異常持續時間(秒)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_metric_rule_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
