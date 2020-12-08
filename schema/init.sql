CREATE TABLE `cms_contents` (
    `id` VARCHAR(50) NOT NULL COMMENT 'content_id',
    `content_type` int NOT NULL COMMENT '数据类型',
    `content_name` VARCHAR(255) NOT NULL COMMENT '内容名称',
    `program` VARCHAR(1024) NOT NULL COMMENT 'program',
    `subject` VARCHAR(1024) NOT NULL COMMENT 'subject',
    `developmental` VARCHAR(1024) NOT NULL COMMENT 'developmental',
    `skills` VARCHAR(1024) NOT NULL COMMENT 'skills',
    `age` VARCHAR(1024) NOT NULL COMMENT 'age',
    `grade` VARCHAR(1024) NOT NULL COMMENT 'grade',
    `keywords` TEXT NULL COMMENT '关键字',
    `description` TEXT NULL COMMENT '描述',
    `thumbnail` TEXT NULL COMMENT '封面',
    `source_type` VARCHAR(255) NULL COMMENT '内容细分类型',
    `data` JSON NULL COMMENT '数据',
    `extra` TEXT NULL COMMENT '附加数据',
    `outcomes` TEXT NULL COMMENT 'Learning outcomes',
    `dir_path` varchar(2048) COMMENT 'Content路径',
    `suggest_time` int NOT NULL COMMENT '建议时间',
    `author` VARCHAR(50) NOT NULL COMMENT '作者id',
    `creator` VARCHAR(50) NOT NULL COMMENT '创建者id',
    `org` VARCHAR(50) NOT NULL COMMENT '所属机构',
    `publish_scope` VARCHAR(50) COMMENT '发布范围',
    `publish_status` VARCHAR(16) NOT NULL COMMENT '状态',
    `self_study` tinyint NOT NULL DEFAULT 0 COMMENT '是否支持自学',
    `draw_activity` tinyint NOT NULL DEFAULT 0 COMMENT '是否支持绘画',
    `reject_reason` VARCHAR(255) COMMENT '拒绝理由',
    `remark` VARCHAR(255) COMMENT '拒绝理由备注',
    `version` INT NOT NULL DEFAULT 0 COMMENT '版本',
    `locked_by` VARCHAR(50) COMMENT '封锁人',
    `source_id` VARCHAR(50) COMMENT 'source_id',
    `latest_id` VARCHAR(50) COMMENT 'latest_id',
    `lesson_type` VARCHAR(100) COMMENT 'lesson_type',
    `create_at` BIGINT NOT NULL COMMENT 'created_at',
    `update_at` BIGINT NOT NULL COMMENT 'updated_at',
    `delete_at` BIGINT NULL COMMENT 'deleted_at',
    PRIMARY KEY (`id`),
    KEY `content_type` (`content_type`),
    KEY `content_author` (`author`),
    KEY `content_org` (`org`),
    KEY `content_publish_status` (`publish_status`),
    KEY `content_source_id` (`source_id`),
    KEY `content_latest_id` (`latest_id`),
    FULLTEXT INDEX `content_name_description_keywords_author_index` (`content_name`, `keywords`, `description`)
) COMMENT '内容表' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
CREATE TABLE IF NOT EXISTS `schedules` (
    `id` varchar(50) NOT NULL COMMENT 'id',
    `title` varchar(100) NOT NULL COMMENT 'title',
    `class_id` varchar(100) NOT NULL COMMENT 'class_id',
    `lesson_plan_id` varchar(100) NULL COMMENT 'lesson_plan_id',
    `org_id` varchar(100) NOT NULL COMMENT 'org_id',
    `subject_id` varchar(100) NULL COMMENT 'subject_id',
    `program_id` varchar(100) NULL COMMENT 'program_id',
    `class_type` varchar(100) NOT NULL COMMENT 'class_type',
    `start_at` bigint(20) NOT NULL COMMENT 'start_at',
    `end_at` bigint(20) NOT NULL COMMENT 'end_at',
    `due_at` bigint(20) DEFAULT NULL COMMENT 'due_at',
    `status` varchar(100) DEFAULT NULL COMMENT 'status',
    `is_all_day` BOOLEAN DEFAULT FALSE COMMENT 'is_all_day',
    `description` varchar(500) DEFAULT NULL COMMENT 'description',
    `attachment` TEXT DEFAULT NULL COMMENT 'attachment',
    `version` bigint(20) DEFAULT 0 COMMENT 'version',
    `repeat_id` varchar(100) DEFAULT NULL COMMENT 'repeat_id',
    `repeat` JSON DEFAULT NULL COMMENT 'repeat',
    `created_id` varchar(100) DEFAULT NULL COMMENT 'created_id',
    `updated_id` varchar(100) DEFAULT NULL COMMENT 'updated_id',
    `deleted_id` varchar(100) DEFAULT NULL COMMENT 'deleted_id',
    `created_at` bigint(20) DEFAULT 0 COMMENT 'created_at',
    `updated_at` bigint(20) DEFAULT 0 COMMENT 'updated_at',
    `delete_at` bigint(20) DEFAULT 0 COMMENT 'delete_at',
    PRIMARY KEY (`id`),
    KEY `schedules_org_id` (`org_id`),
    KEY `schedules_start_at` (`start_at`),
    KEY `schedules_end_at` (`end_at`),
    KEY `schedules_deleted_at` (`delete_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = 'schedules';
CREATE TABLE `learning_outcomes` (
    `id` VARCHAR(50) NOT NULL COMMENT 'outcome_id',
    `ancestor_id` VARCHAR(50) NOT NULL COMMENT 'ancestor_id',
    `shortcode` CHAR(20) NOT NULL COMMENT 'ancestor_id',
    `name` VARCHAR(255) NOT NULL COMMENT 'outcome_name',
    `program` VARCHAR(1024) NOT NULL COMMENT 'program',
    `subject` VARCHAR(1024) NOT NULL COMMENT 'subject',
    `developmental` VARCHAR(1024) NOT NULL COMMENT 'developmental',
    `skills` VARCHAR(1024) NOT NULL COMMENT 'skills',
    `age` VARCHAR(1024) NOT NULL COMMENT 'age',
    `grade` VARCHAR(1024) NOT NULL COMMENT 'grade',
    `keywords` TEXT NULL COMMENT 'keywords',
    `description` TEXT NULL COMMENT 'description',
    `estimated_time` int NOT NULL COMMENT 'estimated_time',
    `author_id` VARCHAR(50) NOT NULL COMMENT 'author_id',
    `author_name` VARCHAR(128) NOT NULL COMMENT 'author_name',
    `organization_id` VARCHAR(50) NOT NULL COMMENT 'organization_id',
    `publish_scope` VARCHAR(50) COMMENT 'publish_scope, default as the organization_id',
    `publish_status` VARCHAR(16) NOT NULL COMMENT 'publish_status',
    `reject_reason` VARCHAR(255) COMMENT 'reject_reason',
    `version` INT NOT NULL DEFAULT 0 COMMENT 'version',
    `assumed` TINYINT NOT NULL DEFAULT 0 COMMENT 'assumed',
    `locked_by` VARCHAR(50) COMMENT 'locked by who',
    `source_id` VARCHAR(50) COMMENT 'source_id',
    `latest_id` VARCHAR(50) COMMENT 'latest_id',
    `create_at` BIGINT NOT NULL COMMENT 'created_at',
    `update_at` BIGINT NOT NULL COMMENT 'updated_at',
    `delete_at` BIGINT NULL COMMENT 'deleted_at',
    PRIMARY KEY (`id`),
    KEY `index_ancestor_id` (`ancestor_id`),
    KEY `index_latest_id` (`latest_id`),
    KEY `index_publish_status` (`publish_status`),
    KEY `index_source_id` (`source_id`),
    FULLTEXT INDEX `fullindex_name_description_keywords_author_shortcode` (
        `name`,
        `keywords`,
        `description`,
        `author_name`,
        `shortcode`
    )
) COMMENT 'outcomes table' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
create table `assessments` (
    `id` varchar(64) not null comment 'id',
    `schedule_id` varchar(64) not null comment 'schedule id',
    `title` varchar(1024) not null comment 'title',
    `program_id` varchar(64) not null comment 'program id',
    `subject_id` varchar(64) not null comment 'subject id',
    `teacher_ids` json not null comment 'teacher id',
    `class_length` int not null comment 'class length (util: minute)',
    `class_end_time` bigint not null comment 'class end time (unix seconds)',
    `complete_time` bigint not null comment 'complete time (unix seconds)',
    `status` varchar(128) not null comment 'status (enum: in_progress, complete)',
    `create_at` bigint not null comment 'create time (unix seconds)',
    `update_at` bigint not null comment 'update time (unix seconds)',
    `delete_at` bigint not null comment 'delete time (unix seconds)',
    primary key (`id`),
    key `assessments_status` (status),
    key `assessments_schedule_id` (schedule_id),
    key `assessments_complete_time` (complete_time)
) comment 'assessment' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
create table assessments_attendances (
    `id` varchar(64) not null comment 'id',
    `assessment_id` varchar(64) not null comment 'assessment id',
    `attendance_id` varchar(64) not null comment 'attendance id',
    `checked` boolean not null comment 'checked',
    primary key (`id`),
    key `assessments_attendances_assessment_id` (`assessment_id`),
    key `assessments_attendances_attendance_id` (`attendance_id`)
) comment 'assessment and attendance map' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
create table assessments_outcomes (
    `id` varchar(64) not null comment 'id',
    `assessment_id` varchar(64) not null comment 'assessment id',
    `outcome_id` varchar(64) not null comment 'outcome id',
    `skip` boolean not null comment 'skip',
    `none_achieved` boolean not null comment 'none achieved',
    primary key (`id`),
    key `assessments_outcomes_assessment_id` (`assessment_id`),
    key `assessments_outcomes_outcome_id` (`outcome_id`)
) comment 'assessment and outcome map' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
create table outcomes_attendances (
    `id` varchar(64) not null comment 'id',
    `assessment_id` varchar(64) not null comment 'assessment id',
    `outcome_id` varchar(64) not null comment 'outcome id',
    `attendance_id` varchar(64) not null comment 'attendance id',
    primary key (`id`),
    key `outcomes_attendances_assessment_id` (`outcome_id`),
    key `outcomes_attendances_outcome_id` (`outcome_id`),
    key `outcomes_attendances_attendance_id` (`attendance_id`)
) comment 'outcome and attendance map' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
CREATE TABLE `cms_folder_items` (
    `id` varchar(50) comment 'id',
    `owner_type` int NOT NULL comment 'folder item owner type',
    `owner` varchar(50) NOT NULL comment 'folder item owner',
    `parent_id` varchar(50) comment 'folder item parent folder id',
    `link` varchar(50) comment 'folder item link',
    `item_type` int NOT NULL comment 'folder item type',
    `dir_path` varchar(2048) NOT NULL comment 'folder item path',
    `editor` varchar(50) NOT NULL comment 'folder item editor',
    `items_count` int NOT NULL comment 'folder item count',
    `name` varchar(256) NOT NULL comment 'folder item name',
    `partition` varchar(256) NOT NULL comment 'folder item partition',
    `thumbnail` text comment 'folder item thumbnail',
    `creator` varchar(50) comment 'folder item creator',
    `create_at` bigint NOT NULL comment 'create time (unix seconds)',
    `update_at` bigint NOT NULL comment 'update time (unix seconds)',
    `delete_at` bigint comment 'delete time (unix seconds)',
    PRIMARY KEY (`id`)
) comment 'cms folder' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
drop index fullindex_name_description_keywords_author_shortcode on learning_outcomes;
create fulltext index fullindex_name_description_keywords_author_shortcode on learning_outcomes(
    `name`,
    `keywords`,
    `description`,
    `author_name`,
    `shortcode`
);
drop index fullindex_name_description_keywords_author_shortcode on learning_outcomes;
alter table learning_outcomes add fulltext index fullindex_name_description_keywords_shortcode(`name`, `keywords`, `description`, `shortcode`);

Create Table: CREATE TABLE `users` (
  `user_id` varchar(64) NOT NULL,
  `user_name` varchar(30) DEFAULT NULL,
  `phone` varchar(24) DEFAULT NULL,
  `email` varchar(80) DEFAULT NULL,
  `secret` varchar(128) DEFAULT NULL,
  `salt` varchar(128) DEFAULT NULL,
  `gender` varchar(8) DEFAULT NULL,
  `birthday` bigint(20) DEFAULT NULL,
  `avatar` text DEFAULT NULL,
  `create_at` bigint(20) DEFAULT '0',
  `update_at` bigint(20) DEFAULT '0',
  `delete_at` bigint(20) DEFAULT '0',
  `create_id` varchar(64) DEFAULT NULL,
  `update_id` varchar(64) DEFAULT NULL,
  `delete_id` varchar(64) DEFAULT NULL,
  `ams_id` varchar(64) DEFAULT NULL,
  PRIMARY KEY (`user_id`),
  UNIQUE KEY `uix_user_phone` (`phone`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE = utf8mb4_unicode_ci;

alter table learning_outcomes
add fulltext index fullindex_name_description_keywords_shortcode(`name`, `keywords`, `description`, `shortcode`);
CREATE TABLE IF NOT EXISTS `user_settings` (
    `id` varchar(50) NOT NULL COMMENT 'id',
    `user_id` varchar(100) NOT NULL COMMENT 'user_id',
    `setting_json` JSON DEFAULT NULL COMMENT 'setting_json',
    PRIMARY KEY (`id`),
    UNIQUE KEY `unique_user_id` (`user_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = 'user_settings';
INSERT INTO `user_settings` (`id`, `user_id`, `setting_json`)
VALUES (
        "default_setting_0",
        "default_setting_0",
        '{"cms_page_size":20}'
    );