CREATE TABLE IF NOT EXISTS `cms_contents` (
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
    `copy_source_id` VARCHAR(50) COMMENT 'copy_source_id',
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
    `is_hidden` BOOLEAN DEFAULT FALSE COMMENT 'is_hidden',
    `is_home_fun` BOOLEAN DEFAULT FALSE COMMENT 'is_home_fun',
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

CREATE TABLE IF NOT EXISTS `schedules_relations` (
    `id` varchar(256) NOT NULL COMMENT  'id',
    `schedule_id` varchar(100) NOT NULL COMMENT  'schedule_id',
    `relation_id` varchar(100) NOT NULL COMMENT  'relation_id',
    `relation_type` varchar(100) DEFAULT NULL COMMENT  'record_type',
    PRIMARY KEY (`id`),
    KEY `idx_schedule_id` (`schedule_id`),
    KEY `idx_relation_id` (`relation_id`),
    KEY `idx_schedule_id_relation_type` (`schedule_id`,`relation_type`)
) COMMENT 'schedules_relations' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;


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
    FULLTEXT INDEX `fullindex_name_description_keywords_shortcode` (
        `name`,
        `keywords`,
        `description`,
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
    `keywords` TEXT NULL COMMENT '关键字',
    `description` TEXT NULL COMMENT '描述',
    `partition` varchar(256) NOT NULL comment 'folder item partition',
    `thumbnail` text comment 'folder item thumbnail',
    `creator` varchar(50) comment 'folder item creator',
    `create_at` bigint NOT NULL comment 'create time (unix seconds)',
    `update_at` bigint NOT NULL comment 'update time (unix seconds)',
    `delete_at` bigint comment 'delete time (unix seconds)',
    PRIMARY KEY (`id`),
    key `folder_name_index` (`name`),
    FULLTEXT INDEX `folder_name_description_keywords_author_index` (`name`, `keywords`, `description`)

) comment 'cms folder' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS  `users` (
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

CREATE TABLE IF NOT EXISTS `user_settings` (
    `id` varchar(50) NOT NULL COMMENT 'id',
    `user_id` varchar(100) NOT NULL COMMENT 'user_id',
    `setting_json` JSON DEFAULT NULL COMMENT 'setting_json',
    PRIMARY KEY (`id`),
    UNIQUE KEY `unique_user_id` (`user_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = 'user_settings';

CREATE TABLE IF NOT EXISTS `cms_authed_contents` (
    `id` VARCHAR(50) NOT NULL COMMENT 'record_id',
    `org_id` VARCHAR(50) NOT NULL COMMENT 'org_id',
    `from_folder_id` varchar(50) COMMENT 'from_folder_id',
    `content_id` VARCHAR(50) NOT NULL COMMENT 'content_id',
    `creator` VARCHAR(50) NOT NULL COMMENT 'creator',
    `duration` INT NOT NULL DEFAULT 0 COMMENT 'duration',
    `create_at` BIGINT NOT NULL COMMENT 'created_at',
    `delete_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'deleted_at',
    PRIMARY KEY (`id`),
    KEY `org_id` (`org_id`),
    KEY `content_id` (`content_id`),
    KEY `creator` (`creator`)
) COMMENT '内容授权记录表' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `cms_shared_folders` (
    `id` VARCHAR(50) NOT NULL COMMENT 'record_id',
    `folder_id` VARCHAR(50) NOT NULL COMMENT 'folder_id',
    `org_id` VARCHAR(50) NOT NULL COMMENT 'org_id',
    `creator` VARCHAR(50) NOT NULL COMMENT 'creator',
    `create_at` BIGINT NOT NULL COMMENT 'created_at',
    `update_at` BIGINT NOT NULL COMMENT 'updated_at',
    `delete_at` BIGINT NOT NULL DEFAULT 0  COMMENT 'deleted_at',
    PRIMARY KEY (`id`),
    KEY `org_id` (`org_id`),
    KEY `folder_id` (`folder_id`),
    KEY `creator` (`creator`)
) COMMENT '文件夹分享记录表' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `organizations_properties` (
    `id` varchar(50) NOT NULL COMMENT 'org_id',
    `type` varchar(200) NOT NULL COMMENT 'type',
    `created_id` varchar(100) DEFAULT NULL COMMENT 'created_id',
    `updated_id` varchar(100) DEFAULT NULL COMMENT 'updated_id',
    `deleted_id` varchar(100) DEFAULT NULL COMMENT 'deleted_id',
    `created_at` bigint(20) DEFAULT 0 COMMENT 'created_at',
    `updated_at` bigint(20) DEFAULT 0 COMMENT 'updated_at',
    `delete_at` bigint(20) DEFAULT 0 COMMENT 'delete_at',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = 'organizations_properties';

CREATE TABLE `ages` (
    `id` varchar(256) NOT NULL COMMENT  'id',
    `name` varchar(255) DEFAULT NULL COMMENT  'name',
    `number` int DEFAULT 0 COMMENT  'number',
    `create_id` varchar(100) DEFAULT NULL COMMENT 'created_id',
    `update_id` varchar(100) DEFAULT NULL COMMENT 'updated_id',
    `delete_id` varchar(100) DEFAULT NULL COMMENT 'deleted_id',
    `create_at` bigint(20) DEFAULT 0 COMMENT 'created_at',
    `update_at` bigint(20) DEFAULT 0 COMMENT 'updated_at',
    `delete_at` bigint(20) DEFAULT 0 COMMENT 'delete_at',
    PRIMARY KEY (`id`),
    KEY `idx_id_delete` (`id`,`delete_at`)
) COMMENT 'ages' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE `lesson_types` (
    `id` varchar(256) NOT NULL COMMENT  'id',
    `name` varchar(255) DEFAULT NULL COMMENT  'name',
    `number` int DEFAULT 0 COMMENT  'number',
    `create_id` varchar(100) DEFAULT NULL COMMENT 'created_id',
    `update_id` varchar(100) DEFAULT NULL COMMENT 'updated_id',
    `delete_id` varchar(100) DEFAULT NULL COMMENT 'deleted_id',
    `create_at` bigint(20) DEFAULT 0 COMMENT 'created_at',
    `update_at` bigint(20) DEFAULT 0 COMMENT 'updated_at',
    `delete_at` bigint(20) DEFAULT 0 COMMENT 'delete_at',
    PRIMARY KEY (`id`),
KEY `idx_id_delete` (`id`,`delete_at`)
) COMMENT 'lesson_types' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE `grades` (
    `id` varchar(256) NOT NULL COMMENT  'id',
    `name` varchar(255) DEFAULT NULL COMMENT  'name',
    `number` int DEFAULT 0 COMMENT  'number',
    `create_id` varchar(100) DEFAULT NULL COMMENT 'created_id',
    `update_id` varchar(100) DEFAULT NULL COMMENT 'updated_id',
    `delete_id` varchar(100) DEFAULT NULL COMMENT 'deleted_id',
    `create_at` bigint(20) DEFAULT 0 COMMENT 'created_at',
    `update_at` bigint(20) DEFAULT 0 COMMENT 'updated_at',
    `delete_at` bigint(20) DEFAULT 0 COMMENT 'delete_at',
    PRIMARY KEY (`id`),
    KEY `idx_id_delete` (`id`,`delete_at`)
) COMMENT 'grades' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE `developmentals` (
    `id` varchar(256) NOT NULL COMMENT  'id',
    `name` varchar(255) DEFAULT NULL COMMENT  'name',
    `number` int DEFAULT 0 COMMENT  'number',
    `create_id` varchar(100) DEFAULT NULL COMMENT 'created_id',
    `update_id` varchar(100) DEFAULT NULL COMMENT 'updated_id',
    `delete_id` varchar(100) DEFAULT NULL COMMENT 'deleted_id',
    `create_at` bigint(20) DEFAULT 0 COMMENT 'created_at',
    `update_at` bigint(20) DEFAULT 0 COMMENT 'updated_at',
    `delete_at` bigint(20) DEFAULT 0 COMMENT 'delete_at',
    PRIMARY KEY (`id`),
    KEY `idx_id_delete` (`id`,`delete_at`)
) COMMENT 'developmentals' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE `class_types` (
    `id` varchar(256) NOT NULL COMMENT  'id',
    `name` varchar(255) DEFAULT NULL COMMENT  'name',
    `number` int DEFAULT 0 COMMENT  'number',
    `create_id` varchar(100) DEFAULT NULL COMMENT 'created_id',
    `update_id` varchar(100) DEFAULT NULL COMMENT 'updated_id',
    `delete_id` varchar(100) DEFAULT NULL COMMENT 'deleted_id',
    `create_at` bigint(20) DEFAULT 0 COMMENT 'created_at',
    `update_at` bigint(20) DEFAULT 0 COMMENT 'updated_at',
    `delete_at` bigint(20) DEFAULT 0 COMMENT 'delete_at',
    PRIMARY KEY (`id`),
    KEY `idx_id_delete` (`id`,`delete_at`)
) COMMENT 'class_types' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE `programs` (
    `id` varchar(256) NOT NULL COMMENT  'id',
    `name` varchar(255) DEFAULT NULL COMMENT  'name',
    `number` int DEFAULT 0 COMMENT  'number',
    `org_type` varchar(100) DEFAULT NULL COMMENT  'org_type',
    `group_name` varchar(100) DEFAULT NULL COMMENT  'group_name',
    `create_id` varchar(100) DEFAULT NULL COMMENT 'created_id',
    `update_id` varchar(100) DEFAULT NULL COMMENT 'updated_id',
    `delete_id` varchar(100) DEFAULT NULL COMMENT 'deleted_id',
    `create_at` bigint(20) DEFAULT 0 COMMENT 'created_at',
    `update_at` bigint(20) DEFAULT 0 COMMENT 'updated_at',
    `delete_at` bigint(20) DEFAULT 0 COMMENT 'delete_at',
    PRIMARY KEY (`id`),
    KEY `idx_id_delete` (`id`,`delete_at`),
    KEY `idx_org_type_delete_at` (`org_type`,`delete_at`),
    KEY `idx_group_name_delete_at` (`group_name`,`delete_at`)
) COMMENT 'programs' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE `subjects` (
    `id` varchar(256) NOT NULL COMMENT  'id',
    `name` varchar(255) DEFAULT NULL COMMENT  'name',
    `number` int DEFAULT 0 COMMENT  'number',
    `create_id` varchar(100) DEFAULT NULL COMMENT 'created_id',
    `update_id` varchar(100) DEFAULT NULL COMMENT 'updated_id',
    `delete_id` varchar(100) DEFAULT NULL COMMENT 'deleted_id',
    `create_at` bigint(20) DEFAULT 0 COMMENT 'created_at',
    `update_at` bigint(20) DEFAULT 0 COMMENT 'updated_at',
    `delete_at` bigint(20) DEFAULT 0 COMMENT 'delete_at',
    PRIMARY KEY (`id`),
    KEY `idx_id_delete` (`id`,`delete_at`)
) COMMENT 'subjects' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE `skills` (
    `id` varchar(256) NOT NULL COMMENT  'id',
    `name` varchar(255) DEFAULT NULL COMMENT  'name',
    `number` int DEFAULT 0 COMMENT  'number',
    `create_id` varchar(100) DEFAULT NULL COMMENT 'created_id',
    `update_id` varchar(100) DEFAULT NULL COMMENT 'updated_id',
    `delete_id` varchar(100) DEFAULT NULL COMMENT 'deleted_id',
    `create_at` bigint(20) DEFAULT 0 COMMENT 'created_at',
    `update_at` bigint(20) DEFAULT 0 COMMENT 'updated_at',
    `delete_at` bigint(20) DEFAULT 0 COMMENT 'delete_at',
    PRIMARY KEY (`id`),
    KEY `idx_id_delete` (`id`,`delete_at`)
) COMMENT 'skills' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE `visibility_settings` (
    `id` varchar(256) NOT NULL COMMENT  'id',
    `name` varchar(255) DEFAULT NULL COMMENT  'name',
    `number` int DEFAULT 0 COMMENT  'number',
    `create_id` varchar(100) DEFAULT NULL COMMENT 'created_id',
    `update_id` varchar(100) DEFAULT NULL COMMENT 'updated_id',
    `delete_id` varchar(100) DEFAULT NULL COMMENT 'deleted_id',
    `create_at` bigint(20) DEFAULT 0 COMMENT 'created_at',
    `update_at` bigint(20) DEFAULT 0 COMMENT 'updated_at',
    `delete_at` bigint(20) DEFAULT 0 COMMENT 'delete_at',
    PRIMARY KEY (`id`),
    KEY `idx_id_delete` (`id`,`delete_at`)
) COMMENT 'visibility_settings' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE IF NOT EXISTS `programs_ages` (
    `id` varchar(50) NOT NULL COMMENT 'id',
    `program_id` varchar(100) NOT NULL COMMENT 'program_id',
    `age_id` varchar(100) NOT NULL COMMENT 'age_id',
    PRIMARY KEY (`id`),
    KEY `idx_program_id` (`program_id`),
    KEY `idx_age_id` (`age_id`)
) ENGINE=InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT='programs_ages';

CREATE TABLE IF NOT EXISTS `programs_developments` (
    `id` varchar(50) NOT NULL COMMENT 'id',
    `program_id` varchar(100) NOT NULL COMMENT 'program_id',
    `development_id` varchar(100) NOT NULL COMMENT 'development_id',
    PRIMARY KEY (`id`),
    KEY `idx_program_id` (`program_id`),
    KEY `idx_development_id` (`development_id`)
) ENGINE=InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT='programs_developments';

CREATE TABLE IF NOT EXISTS `programs_grades` (
    `id` varchar(50) NOT NULL COMMENT 'id',
    `program_id` varchar(100) NOT NULL COMMENT 'program_id',
    `grade_id` varchar(100) NOT NULL COMMENT 'grade_id',
    PRIMARY KEY (`id`),
    KEY `idx_program_id` (`program_id`),
    KEY `idx_grade_id` (`grade_id`)
) ENGINE=InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT='programs_grades';

CREATE TABLE IF NOT EXISTS `developments_skills` (
    `id` varchar(50) NOT NULL COMMENT 'id',
    `program_id` varchar(100) NOT NULL COMMENT 'program_id',
    `development_id` varchar(100) NOT NULL COMMENT 'development_id',
    `skill_id` varchar(100) NOT NULL COMMENT 'skill_id',
    PRIMARY KEY (`id`),
    KEY `idx_program_id` (`program_id`),
    KEY `idx_development_id` (`development_id`),
    KEY `idx_skill_id` (`skill_id`),
    Key `idx_program_develop_skill` (`program_id`,`development_id`,`skill_id`)
) ENGINE=InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT='developments_skills';

CREATE TABLE IF NOT EXISTS `programs_subjects` (
    `id` varchar(50) NOT NULL COMMENT 'id',
    `program_id` varchar(100) NOT NULL COMMENT 'program_id',
    `subject_id` varchar(100) NOT NULL COMMENT 'subject_id',
    PRIMARY KEY (`id`),
    KEY `idx_program_id` (`program_id`),
    KEY `idx_subject_id` (`subject_id`)
) ENGINE=InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT='programs_subjects';

CREATE TABLE IF NOT EXISTS `schedules_feedbacks` (
    `id` varchar(256) NOT NULL COMMENT  'id',
    `schedule_id` varchar(100) NOT NULL DEFAULT "" COMMENT  'schedule_id',
    `user_id` varchar(100) NOT NULL DEFAULT "" COMMENT  'user_id',
    `comment` TEXT DEFAULT NULL COMMENT  'Comment',
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
    `attachment_id` varchar(255) NOT NULL DEFAULT "" COMMENT  'attachment_id',
    `attachment_name` varchar(255) DEFAULT NULL COMMENT  'attachment_name',
    `number` int DEFAULT 0 COMMENT  'number',
    `create_at` bigint(20) DEFAULT 0 COMMENT 'create_at',
    `update_at` bigint(20) DEFAULT 0 COMMENT 'update_at',
    `delete_at` bigint(20) DEFAULT 0 COMMENT 'delete_at',
    PRIMARY KEY (`id`),
    KEY `idx_feedback_id` (`feedback_id`)
) COMMENT 'feedbacks_assignments' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE IF NOT EXISTS `home_fun_studies` (
    `id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'id',
    `schedule_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'schedule id',
    `title` VARCHAR(1024) NOT NULL DEFAULT  '' COMMENT 'title',
    `teacher_ids` JSON NOT NULL COMMENT 'teacher id',
    `student_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'student id',
    `subject_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'subject id',
    `status` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'status (enum: in_progress, complete)',
    `due_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'due at',
    `complete_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'complete at (unix seconds)',
    `latest_feedback_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'latest feedback id',
    `latest_feedback_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'latest feedback at (unix seconds)',
    `assess_feedback_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'assess feedback id',
    `assess_score` INT NOT NULL DEFAULT 0 COMMENT 'score',
    `assess_comment` TEXT NOT NULL COMMENT 'text',
    `create_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'create at (unix seconds)',
    `update_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'update at (unix seconds)',
    `delete_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'delete at (unix seconds)',
    PRIMARY KEY (`id`),
    KEY `home_fun_studies_schedule_id` (schedule_id),
    KEY `home_fun_studies_status` (status),
    KEY `home_fun_studies_latest_feedback_at` (latest_feedback_at),
    KEY `home_fun_studies_complete_at` (complete_at),
    KEY `home_fun_studies_schedule_id_and_student_id` (schedule_id, student_id)
) COMMENT 'home_fun_studies' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;