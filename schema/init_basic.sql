CREATE TABLE `ages` (
`id` varchar(256) NOT NULL COMMENT  'id',
`name` varchar(255) DEFAULT NULL COMMENT  'name',
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
`create_id` varchar(100) DEFAULT NULL COMMENT 'created_id',
`update_id` varchar(100) DEFAULT NULL COMMENT 'updated_id',
`delete_id` varchar(100) DEFAULT NULL COMMENT 'deleted_id',
`create_at` bigint(20) DEFAULT 0 COMMENT 'created_at',
`update_at` bigint(20) DEFAULT 0 COMMENT 'updated_at',
`delete_at` bigint(20) DEFAULT 0 COMMENT 'delete_at',
PRIMARY KEY (`id`),
KEY `idx_id_delete` (`id`,`delete_at`)
) COMMENT 'programs' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE `subjects` (
`id` varchar(256) NOT NULL COMMENT  'id',
`name` varchar(255) DEFAULT NULL COMMENT  'name',
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

-- init data
INSERT INTO `visibility_settings` (`id`,`name`) VALUES ("visibility_settings1","library_label_visibility_school"),("visibility_settings2","library_label_visibility_organization");
INSERT INTO `class_types` (`id`,`name`) VALUES ("OnlineClass","schedule_detail_online_class"),("OfflineClass","schedule_detail_offline_class"),("Homework","schedule_detail_homework"),("Task","schedule_detail_task");
INSERT INTO `lesson_types` (`id`,`name`) VALUES ("1","Test"),("2","Not Test");

INSERT INTO `ages` (`id`,`name`) VALUES ("age0","None Specified");
INSERT INTO `developmentals` (`id`,`name`) VALUES ("developmental0","None Specified");
INSERT INTO `programs` (`id`,`name`) VALUES ("program0","None Specified");
INSERT INTO `skills` (`id`,`name`) VALUES ("skills0","None Specified");
INSERT INTO `subjects` (`id`,`name`) VALUES ("subject0","None Specified");
INSERT INTO `grades` (`id`,`name`) VALUES ("grade0","None Specified");
