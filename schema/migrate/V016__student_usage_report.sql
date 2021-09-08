CREATE TABLE IF NOT EXISTS `classes_assignments_records` (
    `id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
    `class_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'class_id',
    `schedule_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'schedule_id',
    `attendance_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'attendance_id',
    `finish_counts` int(11) NOT NULL DEFAULT 0 COMMENT 'finish counts',
    `schedule_type` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'schedule_type',
    `schedule_start_at` bigint NOT NULL DEFAULT '0' COMMENT 'schedule_start_at',
    `last_end_at` bigint NOT NULL DEFAULT '0' COMMENT 'last_end_at',
    `created_at` bigint DEFAULT '0' COMMENT 'created_at',
    PRIMARY KEY (`id`),
    KEY `index_class_id` (`class_id`),
    KEY `index_attendance_id` (`attendance_id`),
    KEY `index_schedule_id` (`schedule_id`),
    KEY `index_schedule_start_at` (`schedule_start_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='classes_assignments_records'
