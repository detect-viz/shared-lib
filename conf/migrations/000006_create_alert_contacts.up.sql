-- 1. 聯絡人表
CREATE TABLE IF NOT EXISTS alert_contacts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT '聯絡人名稱',
    type ENUM('email', 'teams', 'webhook') NOT NULL COMMENT '通知類型',
    target VARCHAR(255) NOT NULL COMMENT '通知目標',
    severity ENUM('info', 'warn', 'crit') NOT NULL COMMENT '通知等級',
    status ENUM('enabled', 'disabled') NOT NULL DEFAULT 'enabled' COMMENT '啟用狀態',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警聯絡人';

-- 2. 規則與聯絡人關聯表
CREATE TABLE IF NOT EXISTS alert_rule_contacts (
    rule_id VARCHAR(36) NOT NULL COMMENT '規則ID',
    contact_id BIGINT NOT NULL COMMENT '聯絡人ID',
    PRIMARY KEY (rule_id, contact_id),
    FOREIGN KEY (contact_id) REFERENCES alert_contacts(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='規則與聯絡人關聯';

-- 3. 觸發日誌表
CREATE TABLE IF NOT EXISTS trigger_logs (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    uuid VARCHAR(36) NOT NULL,
    timestamp BIGINT NOT NULL COMMENT '觸發時間',
    first_trigger_time BIGINT NOT NULL COMMENT '首次觸發時間',
    rule_id VARCHAR(36) NOT NULL COMMENT '規則ID',
    rule_name VARCHAR(100) NOT NULL COMMENT '規則名稱',
    resource_group VARCHAR(100) NOT NULL COMMENT '資源群組',
    resource_name VARCHAR(100) NOT NULL COMMENT '資源名稱',
    partition_name VARCHAR(100) NULL COMMENT '分區名稱',
    metric VARCHAR(100) NOT NULL COMMENT '監控指標',
    value DECIMAL(10,2) NOT NULL COMMENT '當前值',
    threshold DECIMAL(10,2) NOT NULL COMMENT '閾值',
    unit VARCHAR(20) NOT NULL COMMENT '單位',
    severity ENUM('info', 'warn', 'crit') NOT NULL COMMENT '嚴重程度',
    duration INT NOT NULL COMMENT '持續時間(秒)',
    status VARCHAR(20) NOT NULL COMMENT '狀態',
    notify_status VARCHAR(20) NOT NULL COMMENT '通知狀態',
    silence_start BIGINT NULL COMMENT '靜音開始時間',
    silence_end BIGINT NULL COMMENT '靜音結束時間',
    pending_end BIGINT NULL COMMENT '抑制結束時間',
    resolved_time BIGINT NULL COMMENT '恢復時間',
    labels JSON NULL COMMENT '標籤',
    deleted_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警觸發日誌';

-- 4. 通知日誌表
CREATE TABLE IF NOT EXISTS notification_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    uuid VARCHAR(36) NOT NULL UNIQUE,
    timestamp BIGINT NOT NULL COMMENT '通知時間',
    contact_id BIGINT NOT NULL COMMENT '聯絡人ID',
    contact_name VARCHAR(100) NOT NULL COMMENT '聯絡人名稱',
    contact_type VARCHAR(20) NOT NULL COMMENT '通知類型',
    severity ENUM('info', 'warn', 'crit') NOT NULL COMMENT '嚴重程度',
    subject VARCHAR(255) NOT NULL COMMENT '通知標題',
    file_path VARCHAR(255) NULL COMMENT '通知內容檔案路徑',
    status ENUM('sent', 'failed', 'pending') NOT NULL DEFAULT 'pending' COMMENT '通知狀態',
    sent_at BIGINT NULL COMMENT '發送時間',
    error TEXT NULL COMMENT '錯誤訊息',
    notify_retry INT NOT NULL DEFAULT 0 COMMENT '重試次數',
    retry_deadline BIGINT NOT NULL COMMENT '重試截止時間',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知日誌';

-- 5. 通知與觸發日誌關聯表
CREATE TABLE IF NOT EXISTS notification_trigger_logs (
    notification_id BIGINT NOT NULL,
    trigger_log_id BIGINT NOT NULL,
    PRIMARY KEY (notification_id, trigger_log_id),
    FOREIGN KEY (notification_id) REFERENCES notification_logs(id),
    FOREIGN KEY (trigger_log_id) REFERENCES trigger_logs(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知與觸發日誌關聯'; 