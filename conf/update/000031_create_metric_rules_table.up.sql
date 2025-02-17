CREATE TABLE `metric_rules` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL,
  `metric_name` varchar(50) NOT NULL,
  `unit` varchar(20)  NOT NULL,
  `operator` varchar(10) NOT NULL,
  `type` varchar(20) NOT NULL,
  `check_type` varchar(20) NOT NULL,
  `default_threshold` double NOT NULL,
  `default_duration` int DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_metric_rule_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;