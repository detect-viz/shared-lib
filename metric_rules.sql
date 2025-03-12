CREATE TABLE `metric_rules` (
  `uid` varchar(50) NOT NULL COMMENT '告警規則ID',
  `name` varchar(100) NOT NULL COMMENT '告警規則名稱',
  `category` varchar(50) NOT NULL COMMENT '類別 (cpu/memory/disk/network/database/filesystem)',
  `check_type` ENUM('absolute','amplitude') NOT NULL COMMENT '檢查類型: 絕對值 / 變化幅度',
  `metric_raw_name` varchar(50) NOT NULL COMMENT '監控指標名稱',
  `metric_display_name` varchar(100) NOT NULL COMMENT '監控指標顯示名稱',
  `metric_pattern` varchar(255) NOT NULL COMMENT '監控指標顯示名稱',
  `raw_unit` varchar(20) NOT NULL COMMENT '原始單位 (bytes, %, MB, GB)',
  `display_unit` varchar(20) NOT NULL COMMENT '顯示單位 (GB, MB/s, %)',
  `scale` DOUBLE NOT NULL DEFAULT 1.0 COMMENT '轉換倍率: raw_value × scale = display_value',
  `duration` INT NOT NULL COMMENT '異常持續時間(秒)',
  `operator` ENUM('gt','ge','lt','le') NOT NULL COMMENT '比較運算符: greater_than, less_than',
  `info_threshold` DOUBLE DEFAULT NULL COMMENT '資訊等級閾值 (可選)',
  `warn_threshold` DOUBLE DEFAULT NULL COMMENT '警告等級閾值',
  `crit_threshold` DOUBLE DEFAULT NULL COMMENT '嚴重等級閾值',
  PRIMARY KEY (`uid`),
  UNIQUE KEY `uk_metric_rule_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

  