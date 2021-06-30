ALTER TABLE assessments CHANGE COLUMN `type` `type` VARCHAR(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'assessment type';

ALTER TABLE milestones CHANGE COLUMN `name` `name` VARCHAR(200) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'name';

CREATE TABLE IF NOT EXISTS `users` (
  `user_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_name` varchar(30) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `phone` varchar(24) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `email` varchar(80) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `secret` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `salt` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `gender` varchar(8) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `birthday` bigint(20) DEFAULT NULL,
  `avatar` text COLLATE utf8mb4_unicode_ci,
  `create_at` bigint(20) DEFAULT '0',
  `update_at` bigint(20) DEFAULT '0',
  `delete_at` bigint(20) DEFAULT '0',
  `create_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `update_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `delete_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `ams_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`user_id`),
  UNIQUE KEY `uix_user_phone` (`phone`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
