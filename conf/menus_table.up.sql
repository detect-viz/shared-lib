INSERT INTO `sub_menus` (`id`, `router_path`, `menu_type`, `icon`, `title`, `sort`, `parent_id`)
VALUES
	(13, 'setting/server-tag', 'internal', 'k-i-group', '伺服器標籤', 2, 4),
	(16, 'alarm/status', 'internal', 'k-i-inherited', '當前告警', 1, 6),
	(17, 'alarm/history', 'internal', 'k-i-chart-area-stacked', '歷史告警', 2, 6),
	(18, 'alarm/list', 'internal', 'k-i-group-collection', '規則列表', 3, 6),
	(19, 'alarm/alarm-setting-list', 'internal', 'k-i-user', '通知管道', 4, 6),
	(20, 'alarm/inhibition-rule', 'internal', 'k-i-volume-mute', '告警抑制', 5, 6),
	(21, 'alarm/alarm-rule-label', 'internal', 'k-i-delicious', '規則標籤', 6, 6);