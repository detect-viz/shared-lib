INSERT INTO `resources` (`name`, `resource_group_id`, `created_at`, `updated_at`, `deleted_at`)
VALUES
	('AL2SUB', '3', NULL, NULL, NULL),
	('bimap-ipoc', '3', '1735803001', '1735803001', NULL),
	('SAPLinuxAP01', '3', NULL, NULL, NULL),
	('SRVECCDV01', '3', NULL, NULL, NULL),
	('SRVIPOCNEW', '3', NULL, NULL, NULL);


INSERT INTO metric_rules (name, metric_name, unit, operator, type, check_type, default_threshold, default_duration) VALUES
('CPU user 使用率過高', 'user_usage', '%', '>', 'cpu', 'absolute', 90.0, 300),
('記憶體使用率過高', 'mem_usage', '%', '>', 'memory', 'absolute', 85.0, 300),
('磁碟使用率過高', 'fs_usage', '%', '>', 'filesystem', 'absolute', 90.0, 300),
('網路流量異常', 'recv_bytes', 'Mbps', '>', 'network', 'absolute', 1000.0, 300),
('CPU system 變化幅度過大', 'system_usage', '%', '>', 'cpu', 'amplitude', 50.0, 600),
('記憶體變化幅度過大', 'mem_usage', '%', '>', 'memory', 'amplitude', 40.0, 600),
('磁碟 busy 變化過快', 'busy', '%', '>', 'disk', 'amplitude', 60.0, 600),
('網路流量突增', 'send_bytes', 'Mbps', '>', 'network', 'amplitude', 200.0, 600);
INSERT INTO `alert_rules` (`realm_name`, `id`, `name`, `enabled`, `is_joint`, `resource_group_id`, `metric_rule_id`, `info_threshold`, `warn_threshold`, `crit_threshold`, `duration`, `silence_period`, `times`, `created_at`, `updated_at`, `deleted_at`)
VALUES
	('master', '1', 'CPU 使用率過高', '1', '0', '3', '1', '70', '85', '90', '300', '1h', '3', NULL, NULL, NULL),
	('master', '2', '記憶體使用率過高', '1', '0', '3', '2', '60', '75', '85', '300', '1h', '3', NULL, NULL, NULL),
	('master', '3', '磁碟使用率過高', '1', '0', '3', '3', '80', '85', '90', '300', '1h', '3', NULL, NULL, NULL),
	('master', '4', '網路流量異常', '1', '0', '3', '4', '500', '800', '1000', '300', '1h', '3', NULL, NULL, NULL);

INSERT INTO contacts (id, realm_name, name, type, enabled, details)
VALUES
    (1, 'master', 'Admin Email', 'email', 1, '{"email": "admin@example.com"}'),
    (2, 'master', 'Line Channel', 'line', 1, '{"url": "https://notify-bot.line.me/","user_id": "cmt3Ss2pIjnXr7mARzOSwdB04t89/1O/w1cDnyilFU=","channel_token": "s9Bao6VivhqM1u6Tap1fBSvIkDXVqbxks9nZG4BAdrifTl4TdGDwBhZXSN/G+EDRAktX+REGJsUXFVeZZsiIOJ8GrhnR2U0QRBwZaRBjThZkRJtTD5/2EQOxCyFpjdFB"}');
INSERT INTO `alert_rule_contacts` (`alert_rule_id`, `alert_contact_id`)
VALUES
	('1', '1'),
	('1', '2'),
	('2', '1'),
	('2', '2'),
	('3', '1'),
	('3', '2'),
	('4', '1'),
	('4', '2');


INSERT INTO `alert_contact_severities` (`alert_contact_id`, `severity`)
VALUES
	('1', 'info'),
	('1', 'warn'),
	('1', 'crit'),
	('2', 'info'),
	('2', 'warn'),
	('2', 'crit');
	

INSERT INTO alert_rule_details (realm_name, alert_rule_id, resource_name, partition_name)
VALUES
    ('master', 1, 'SRVECCDV01', 'total'),
    ('master', 1, 'SRVIPOCNEW', 'total'),
    ('master', 1, 'SAPLinuxAP01', 'total'),
    ('master', 1, 'AL2SUB', 'total'),
    ('master', 2, 'SRVECCDV01', null),
    ('master', 2, 'SRVIPOCNEW', null),
    ('master', 2, 'SAPLinuxAP01', null),
    ('master', 2, 'AL2SUB', null);

INSERT INTO alert_rule_details (realm_name, alert_rule_id, resource_name, partition_name)
VALUES
    ('master', 1, 'SRVECCDV01', 'cpu0'),
    ('master', 2, 'SRVIPOCNEW', null);


INSERT INTO `labels` (`realm_name`, `id`, `key`, `value`, `created_at`, `updated_at`)
VALUES
	('master', '1', '環境標籤', '正式區', NULL, NULL),
	('master', '2', '環境標籤', '測試區', NULL, NULL);

INSERT INTO `alert_rule_labels` (`realm_name`, `rule_id`, `label_id`, `created_at`, `updated_at`)
VALUES
	('master', '1', '1', NULL, NULL),
	('master', '2', '2', NULL, NULL);
    
