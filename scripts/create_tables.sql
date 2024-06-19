-- ----------------------------
-- Table structure for users
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
`id` bigint(20) NOT NULL AUTO_INCREMENT,
`uuid` varchar(255) NOT NULL DEFAULT '',
`avatar` varchar(255) NOT NULL DEFAULT '',
`username` varchar(255) NOT NULL DEFAULT '',
`pass_hash` varchar(255) NOT NULL DEFAULT '',
`user_email` varchar(255) NOT NULL DEFAULT '',
`wallet_address` varchar(255) NOT NULL DEFAULT '',
`role` tinyint(4) NOT NULL DEFAULT '0',
`allocate_storage` int(1) NOT NULL DEFAULT '0',
`created_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
`updated_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
`deleted_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
`project_id` int(20) NOT NULL DEFAULT '0',
`referral_code` varchar(64) NOT NULL DEFAULT '',
`referrer` varchar(64) NOT NULL DEFAULT '',
`referrer_user_id` varchar(255) NOT NULL DEFAULT '',
PRIMARY KEY (`id`),
UNIQUE KEY `uniq_username` (`username`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=2058 DEFAULT CHARSET=utf8mb4;


-- ----------------------------
-- Table structure for location_cn
-- ----------------------------
DROP TABLE IF EXISTS `location_cn`;
CREATE TABLE `location_cn` (
`id` bigint(20) NOT NULL AUTO_INCREMENT,
`ip` varchar(28) NOT NULL DEFAULT '',
`continent` varchar(28) NOT NULL DEFAULT '',
`country` varchar(128) NOT NULL DEFAULT '',
`province` varchar(128) NOT NULL DEFAULT '',
`city` varchar(128) NOT NULL DEFAULT '',
`longitude` varchar(28) NOT NULL DEFAULT '',
`area_code` varchar(28) NOT NULL DEFAULT '',
`latitude` varchar(28) NOT NULL DEFAULT '',
`isp` varchar(256) NOT NULL DEFAULT '',
`zip_code` varchar(28) NOT NULL DEFAULT '',
`elevation` varchar(28) NOT NULL DEFAULT '',
`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
PRIMARY KEY (`id`),
UNIQUE KEY `uniq_uuid` (`ip`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=42497 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for location_en
-- ----------------------------
DROP TABLE IF EXISTS `location_en`;
CREATE TABLE `location_en` (
`id` bigint(20) NOT NULL AUTO_INCREMENT,
`ip` varchar(28) NOT NULL DEFAULT '',
`continent` varchar(28) NOT NULL DEFAULT '',
`country` varchar(128) NOT NULL DEFAULT '',
`province` varchar(128) NOT NULL DEFAULT '',
`city` varchar(128) NOT NULL DEFAULT '',
`longitude` varchar(28) NOT NULL DEFAULT '',
`area_code` varchar(28) NOT NULL DEFAULT '',
`latitude` varchar(28) NOT NULL DEFAULT '',
`isp` varchar(256) NOT NULL DEFAULT '',
`zip_code` varchar(28) NOT NULL DEFAULT '',
`elevation` varchar(28) NOT NULL DEFAULT '',
`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
PRIMARY KEY (`id`),
UNIQUE KEY `uniq_uuid` (`ip`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=42492 DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `project`;
CREATE TABLE `project` (
`id` bigint(20) NOT NULL AUTO_INCREMENT,
`user_id` varchar(128) NOT NULL DEFAULT '',
`project_id` varchar(128) NOT NULL DEFAULT '',
`name` varchar(128) NOT NULL DEFAULT '',
`area_id` varchar(128) NOT NULL DEFAULT '',
`region` varchar(128) NOT NULL DEFAULT '',
`bundle_url` text NOT NULL,
`status` varchar(128) NOT NULL DEFAULT '',
`replicas` bigint(20) NOT NULL DEFAULT 0,
`cpu_cores` int NOT NULL DEFAULT 0,
`memory` float NOT NULL DEFAULT 0,
`expiration` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
`node_ids` varchar(256) NOT NULL DEFAULT '',
`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
PRIMARY KEY (`id`),
UNIQUE KEY `uniq_project_id` (`project_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;
