INSERT INTO `resource_groups` (`realm_name`, `type`, `id`, `name`,`realm_groups`)
VALUES
	('master', 'host', '1', 'group-1','[\"8b907375-2b56-4eb0-a0d7-3c2e196fd74e\"]'),
    ('master', 'host', '2', 'group-2','[\"8b907375-2b56-4eb0-a0d7-3c2e196fd74e\"]');

INSERT INTO `resources` (`name`, `resource_group_id`)
VALUES
	('AL2SUB', '1'),
	('FLOW-WEB', '1'),
	('FLOWSERVER', '1'),
	('SRVECCDV01', '1'),
	('SRVECCPD02', '1'),
	('SRVWNCPRD02', '1'),
	('SRVWNCPRD01', '1');