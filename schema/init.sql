CREATE TABLE `kidsloop2`.`assets` (
    `id` VARCHAR(50) NOT NULL,
    `name` VARCHAR(256) NOT NULL,
    `program` VARCHAR(50) NOT NULL,
    `subject` VARCHAR(50) NOT NULL,
    `developmental` VARCHAR(50) NOT NULL,
    `skills` VARCHAR(50) NOT NULL,
    `age` VARCHAR(50) NOT NULL,
    `keywords` TEXT NULL,
    `description` TEXT NULL,
    `thumbnail` TEXT NOT NULL,
    `size` BIGINT NOT NULL DEFAULT 0,
    `resource` TEXT NOT NULL,
    `author` VARCHAR(50) NOT NULL,
    `author_name` VARCHAR(128) NOT NULL,
    `org` VARCHAR(50) NOT NULL,
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    `deleted_at` DATETIME NULL,
    PRIMARY KEY (`id`),
    FULLTEXT INDEX `name_description_keywords_author_index` (`name`, `keywords`, `description`, `author_name`) WITH PARSER ngram
);

CREATE TABLE IF NOT EXISTS `kidsloop2`.`schedules` (
  `id` varchar(50) NOT NULL COMMENT 'id',
  `title` varchar(100) NOT NULL COMMENT 'title',
  `class_id` varchar(100) NOT NULL COMMENT 'class_id',
  `lesson_plan_id` varchar(100) NOT NULL COMMENT 'lesson_plan_id',
  `org_id` varchar(100) NOT NULL COMMENT 'org_id',
  `subject_id` varchar(100) NOT NULL COMMENT 'subject_id',
  `program_id` varchar(100) NOT NULL COMMENT 'program_id',
  `class_type` varchar(100) NOT NULL COMMENT 'class_type',
  `start_at` bigint(20) NOT NULL COMMENT 'start_at',
  `end_at` bigint(20) NOT NULL COMMENT 'end_at',
  `due_at` bigint(20) DEFAULT NULL COMMENT 'due_at',
  `description` varchar(500) DEFAULT NULL COMMENT 'description',
  `attachment_url` varchar(500) DEFAULT NULL COMMENT 'attachment_url',
  `version` bigint(20) DEFAULT 0 COMMENT 'version',
  `repeat_id` varchar(100) DEFAULT NULL COMMENT 'repeat_id',
  `repeat` varchar(500) DEFAULT NULL COMMENT 'repeat',
  `created_id` varchar(100) DEFAULT NULL COMMENT 'created_id',
  `updated_id` varchar(100) DEFAULT NULL COMMENT 'updated_id',
  `deleted_id` varchar(100) DEFAULT NULL COMMENT 'deleted_id',
  `created_at` bigint(20) DEFAULT 0 COMMENT 'created_at',
  `updated_at` bigint(20) DEFAULT 0 COMMENT 'updated_at',
  `deleted_at` bigint(20) DEFAULT 0 COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  KEY `schedules_org_id` (`org_id`),
  KEY `schedules_start_at` (`start_at`),
  KEY `schedules_end_at` (`end_at`),
  KEY `schedules_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='schedules';

CREATE TABLE IF NOT EXISTS `kidsloop2`.`schedules_teachers` (
  `id` varchar(50) NOT NULL COMMENT 'id',
  `teacher_id` varchar(100) NOT NULL COMMENT 'teacher_id',
  `schedule_id` varchar(100) NOT NULL COMMENT 'schedule_id',
  `deleted_at` bigint(20) DEFAULT 0 COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  KEY `schedules_teacher_id` (`teacher_id`),
  KEY `schedules_schedule_id` (`schedule_id`),
  KEY `schedules_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='schedules_teachers';


