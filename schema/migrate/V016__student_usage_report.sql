CREATE TABLE IF NOT EXISTS `classes_assignments_records` (
    `id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
    `class_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'class_id',
    `schedule_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'schedule_id',
    `attendance_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'attendance_id',
    `finish_counts` int(11) NOT NULL DEFAULT 0 COMMENT 'finish counts',
    `schedule_type` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'schedule_type',
    `schedule_start_at` bigint NOT NULL DEFAULT '0' COMMENT 'schedule_start_at',
    `last_end_at` bigint NOT NULL DEFAULT '0' COMMENT 'last_end_at',
    `create_at` bigint DEFAULT '0' COMMENT 'create_at',
    PRIMARY KEY (`id`),
    KEY `index_class_id` (`class_id`),
    KEY `index_attendance_id` (`attendance_id`),
    KEY `index_schedule_id` (`schedule_id`),
    KEY `index_schedule_start_at` (`schedule_start_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='classes_assignments_records';

CREATE TABLE IF NOT EXISTS `student_usage_records` (
    `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
    `class_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' ,
    `room_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' ,
    `lesson_material_url` varchar(2100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' ,
    `content_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' ,
    `action_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' ,
    `timestamp` bigint COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 0 ,
    `student_user_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' ,
    `student_email` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' ,
    `student_name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' ,
    `lesson_material_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' ,
    `lesson_plan_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' ,
    `schedule_start_at` bigint COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 0 ,
    `class_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' ,
    PRIMARY KEY (`id`),
    KEY `index_student_lesson_plan_lesson_material_class_content_type` (`student_user_id`,`lesson_plan_id`,`lesson_material_id`,`class_id`,`content_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
