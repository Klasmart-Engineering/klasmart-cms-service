ALTER TABLE `schedules` ADD COLUMN is_hidden BOOLEAN DEFAULT FALSE COMMENT 'is hidden';
ALTER TABLE `schedules` ADD COLUMN is_home_fun BOOLEAN DEFAULT FALSE COMMENT 'is home fun';

CREATE TABLE IF NOT EXISTS `schedules_feedbacks` (
    `id` varchar(256) NOT NULL COMMENT  'id',
    `schedule_id` varchar(100) NOT NULL DEFAULT "" COMMENT  'schedule_id',
    `user_id` varchar(100) NOT NULL DEFAULT "" COMMENT  'user_id',
    `Comment` TEXT DEFAULT NULL COMMENT  'Comment',
    `create_at` bigint(20) DEFAULT 0 COMMENT 'create_at',
    `update_at` bigint(20) DEFAULT 0 COMMENT 'update_at',
    `delete_at` bigint(20) DEFAULT 0 COMMENT 'delete_at',
    PRIMARY KEY (`id`),
    KEY `idx_schedule_id` (`schedule_id`),
    KEY `idx_user_id` (`user_id`)
) COMMENT 'schedules_feedbacks' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE IF NOT EXISTS `feedbacks_assignments` (
    `id` varchar(256) NOT NULL COMMENT  'id',
    `feedback_id` varchar(100) NOT NULL DEFAULT "" COMMENT  'feedback_id',
    `attachment_id` varchar(500) NOT NULL DEFAULT "" COMMENT  'attachment_id',
    `attachment_name` varchar(500) DEFAULT NULL COMMENT  'attachment_name',
    `number` int DEFAULT 0 COMMENT  'number',
    `create_at` bigint(20) DEFAULT 0 COMMENT 'create_at',
    `update_at` bigint(20) DEFAULT 0 COMMENT 'update_at',
    `delete_at` bigint(20) DEFAULT 0 COMMENT 'delete_at',
    PRIMARY KEY (`id`),
    KEY `idx_feedback_id` (`feedback_id`)
) COMMENT 'feedbacks_assignments' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;
