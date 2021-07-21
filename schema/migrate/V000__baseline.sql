CREATE TABLE IF NOT EXISTS `ages` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'name',
  `number` int(11) DEFAULT '0' COMMENT 'number',
  `create_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `update_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `delete_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `create_at` bigint(20) DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint(20) DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT '0' COMMENT 'delete_at',
  PRIMARY KEY (`id`),
  KEY `idx_id_delete` (`id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='ages';


CREATE TABLE IF NOT EXISTS `assessments` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `schedule_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'schedule id',
  `title` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'title',
  `program_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'DEPRECATED: program id',
  `subject_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'DEPRECATED: subject id',
  `teacher_ids` json DEFAULT NULL COMMENT 'DEPRECATED: teacher ids',
  `class_length` int(11) NOT NULL COMMENT 'class length (util: minute)',
  `class_end_time` bigint(20) NOT NULL COMMENT 'class end time (unix seconds)',
  `complete_time` bigint(20) NOT NULL COMMENT 'complete time (unix seconds)',
  `status` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'status (enum: in_progress, complete)',
  `create_at` bigint(20) NOT NULL COMMENT 'create time (unix seconds)',
  `update_at` bigint(20) NOT NULL COMMENT 'update time (unix seconds)',
  `delete_at` bigint(20) NOT NULL COMMENT 'delete time (unix seconds)',
  `type` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'assessment type',
  PRIMARY KEY (`id`),
  KEY `assessments_status` (`status`),
  KEY `assessments_schedule_id` (`schedule_id`),
  KEY `assessments_complete_time` (`complete_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='assessment';


CREATE TABLE IF NOT EXISTS `assessments_attendances` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `assessment_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'assessment id',
  `attendance_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'attendance id',
  `checked` tinyint(1) NOT NULL COMMENT 'checked',
  `origin` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'origin',
  `role` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'role',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_assessments_attendances_assessment_id_attendance_id_role` (`assessment_id`,`attendance_id`,`role`),
  KEY `assessments_attendances_assessment_id` (`assessment_id`),
  KEY `assessments_attendances_attendance_id` (`attendance_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='assessment and attendance map';


CREATE TABLE IF NOT EXISTS `assessments_contents` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'id',
  `assessment_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'assessment id',
  `content_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'content id',
  `content_name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'content name',
  `content_type` int(11) NOT NULL DEFAULT '0' COMMENT 'content type',
  `content_comment` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'content comment',
  `checked` tinyint(1) NOT NULL DEFAULT '1' COMMENT 'checked',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_assessments_contents_assessment_id_content_id` (`assessment_id`,`content_id`),
  KEY `idx_assessments_contents_assessment_id` (`assessment_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='assessment and outcome map';


CREATE TABLE IF NOT EXISTS `assessments_contents_outcomes` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'id',
  `assessment_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'assessment id',
  `content_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'content id',
  `outcome_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'outcome id',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_assessments_contents_outcomes` (`assessment_id`,`content_id`,`outcome_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='assessment content and outcome map';


CREATE TABLE IF NOT EXISTS `assessments_outcomes` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `assessment_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'assessment id',
  `outcome_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'outcome id',
  `skip` tinyint(1) NOT NULL COMMENT 'skip',
  `none_achieved` tinyint(1) NOT NULL COMMENT 'none achieved',
  `checked` tinyint(1) NOT NULL DEFAULT '1' COMMENT 'checked',
  PRIMARY KEY (`id`),
  KEY `assessments_outcomes_assessment_id` (`assessment_id`),
  KEY `assessments_outcomes_outcome_id` (`outcome_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='assessment and outcome map';


CREATE TABLE IF NOT EXISTS `class_types` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'name',
  `number` int(11) DEFAULT '0' COMMENT 'number',
  `create_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `update_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `delete_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `create_at` bigint(20) DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint(20) DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT '0' COMMENT 'delete_at',
  PRIMARY KEY (`id`),
  KEY `idx_id_delete` (`id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='class_types';


CREATE TABLE IF NOT EXISTS `cms_authed_contents` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'record_id',
  `org_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'org_id',
  `from_folder_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'from_folder_id',
  `content_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'content_id',
  `creator` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'creator',
  `duration` int(11) NOT NULL DEFAULT '0' COMMENT 'duration',
  `create_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'created_at',
  `delete_at` bigint(20) DEFAULT '0' COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  KEY `org_id` (`org_id`),
  KEY `content_id` (`content_id`),
  KEY `creator` (`creator`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='内容授权记录表';


CREATE TABLE IF NOT EXISTS `cms_content_properties` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `content_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'content id',
  `property_type` int(11) NOT NULL DEFAULT '0' COMMENT 'property type',
  `property_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'property id',
  `sequence` int(11) NOT NULL DEFAULT '0' COMMENT 'sequence',
  PRIMARY KEY (`id`),
  KEY `cms_content_properties_content_id_idx` (`content_id`),
  KEY `cms_content_properties_property_type_idx` (`property_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='cms content properties';


CREATE TABLE IF NOT EXISTS `cms_content_visibility_settings` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `content_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'content id',
  `visibility_setting` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'visibility setting',
  PRIMARY KEY (`id`),
  KEY `cms_content_visibility_settings_content_id_idx` (`content_id`),
  KEY `cms_content_visibility_settings_visibility_settings_idx` (`visibility_setting`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='cms content visibility settings';


CREATE TABLE IF NOT EXISTS `cms_contents` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'content_id',
  `content_type` int(11) NOT NULL COMMENT '数据类型',
  `content_name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '内容名称',
  `program` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'program',
  `subject` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'subject',
  `developmental` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'developmental',
  `skills` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'skills',
  `age` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'age',
  `grade` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'grade',
  `keywords` text COLLATE utf8mb4_unicode_ci COMMENT '关键字',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '描述',
  `thumbnail` text COLLATE utf8mb4_unicode_ci COMMENT '封面',
  `source_type` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '内容细分类型',
  `data` json DEFAULT NULL COMMENT '数据',
  `extra` text COLLATE utf8mb4_unicode_ci COMMENT '附加数据',
  `outcomes` text COLLATE utf8mb4_unicode_ci COMMENT 'Learning outcomes',
  `dir_path` varchar(2048) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Content路径',
  `suggest_time` int(11) NOT NULL COMMENT '建议时间',
  `author` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '作者id',
  `creator` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '创建者id',
  `org` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '所属机构',
  `publish_scope` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '发布范围',
  `publish_status` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '状态',
  `self_study` tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否支持自学',
  `draw_activity` tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否支持绘画',
  `reject_reason` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '拒绝理由',
  `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '拒绝理由备注',
  `version` int(11) NOT NULL DEFAULT '0' COMMENT '版本',
  `locked_by` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '封锁人',
  `source_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'source_id',
  `copy_source_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'copy_source_id',
  `latest_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'latest_id',
  `lesson_type` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'lesson_type',
  `create_at` bigint(20) NOT NULL COMMENT 'created_at',
  `update_at` bigint(20) NOT NULL COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT NULL COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  KEY `content_type` (`content_type`),
  KEY `content_author` (`author`),
  KEY `content_org` (`org`),
  KEY `content_publish_status` (`publish_status`),
  KEY `content_source_id` (`source_id`),
  KEY `content_latest_id` (`latest_id`),
  FULLTEXT KEY `content_name_description_keywords_author_index` (`content_name`,`keywords`,`description`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='内容表';


CREATE TABLE IF NOT EXISTS `cms_folder_items` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `owner_type` int(11) NOT NULL COMMENT 'folder item owner type',
  `owner` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'folder item owner',
  `parent_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'folder item parent folder id',
  `link` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'folder item link',
  `item_type` int(11) NOT NULL COMMENT 'folder item type',
  `dir_path` varchar(2048) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'folder item path',
  `editor` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'folder item editor',
  `items_count` int(11) NOT NULL COMMENT 'folder item count',
  `name` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'folder item name',
  `partition` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'folder item partition',
  `thumbnail` text COLLATE utf8mb4_unicode_ci COMMENT 'folder item thumbnail',
  `creator` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'folder item creator',
  `create_at` bigint(20) NOT NULL COMMENT 'create time (unix seconds)',
  `update_at` bigint(20) NOT NULL COMMENT 'update time (unix seconds)',
  `delete_at` bigint(20) DEFAULT NULL COMMENT 'delete time (unix seconds)',
  `keywords` text COLLATE utf8mb4_unicode_ci COMMENT '关键字',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '描述',
  PRIMARY KEY (`id`),
  FULLTEXT KEY `folder_name_description_keywords_author_index` (`name`,`keywords`,`description`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='cms folder';


CREATE TABLE IF NOT EXISTS `cms_shared_folders` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'record_id',
  `folder_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'folder_id',
  `org_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'org_id',
  `creator` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'creator',
  `create_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT '0' COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  KEY `org_id` (`org_id`),
  KEY `folder_id` (`folder_id`),
  KEY `creator` (`creator`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件夹分享记录表';


CREATE TABLE IF NOT EXISTS `developmentals` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'name',
  `number` int(11) DEFAULT '0' COMMENT 'number',
  `create_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `update_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `delete_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `create_at` bigint(20) DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint(20) DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT '0' COMMENT 'delete_at',
  PRIMARY KEY (`id`),
  KEY `idx_id_delete` (`id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='developmentals';


CREATE TABLE IF NOT EXISTS `developments_skills` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `program_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'program_id',
  `development_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'development_id',
  `skill_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'skill_id',
  PRIMARY KEY (`id`),
  KEY `idx_program_id` (`program_id`),
  KEY `idx_development_id` (`development_id`),
  KEY `idx_skill_id` (`skill_id`),
  KEY `idx_program_develop_skill` (`program_id`,`development_id`,`skill_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='developments_skills';


CREATE TABLE IF NOT EXISTS `feedbacks_assignments` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `feedback_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'feedback_id',
  `attachment_id` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'attachment_id',
  `attachment_name` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'attachment_name',
  `number` int(11) DEFAULT '0' COMMENT 'number',
  `create_at` bigint(20) DEFAULT '0' COMMENT 'create_at',
  `update_at` bigint(20) DEFAULT '0' COMMENT 'update_at',
  `delete_at` bigint(20) DEFAULT '0' COMMENT 'delete_at',
  PRIMARY KEY (`id`),
  KEY `idx_feedback_id` (`feedback_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='feedbacks_assignments';


CREATE TABLE IF NOT EXISTS `grades` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'name',
  `number` int(11) DEFAULT '0' COMMENT 'number',
  `create_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `update_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `delete_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `create_at` bigint(20) DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint(20) DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT '0' COMMENT 'delete_at',
  PRIMARY KEY (`id`),
  KEY `idx_id_delete` (`id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='grades';


CREATE TABLE IF NOT EXISTS `home_fun_studies` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'id',
  `schedule_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'schedule id',
  `title` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'title',
  `teacher_ids` json NOT NULL COMMENT 'teacher id',
  `student_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'student id',
  `subject_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'subject id',
  `status` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'status (enum: in_progress, complete)',
  `due_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'due at',
  `complete_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'complete at (unix seconds)',
  `latest_feedback_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'latest feedback id',
  `latest_feedback_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'latest feedback at (unix seconds)',
  `assess_feedback_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'assess feedback id',
  `assess_score` int(11) NOT NULL DEFAULT '0' COMMENT 'score',
  `assess_comment` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'text',
  `create_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'create at (unix seconds)',
  `update_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'update at (unix seconds)',
  `delete_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'delete at (unix seconds)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_home_fun_studies_schedule_id_and_student_id` (`schedule_id`,`student_id`),
  KEY `idx_home_fun_studies_schedule_id` (`schedule_id`),
  KEY `idx_home_fun_studies_status` (`status`),
  KEY `idx_home_fun_studies_latest_feedback_at` (`latest_feedback_at`),
  KEY `idx_home_fun_studies_complete_at` (`complete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='home_fun_studies';


CREATE TABLE IF NOT EXISTS `learning_outcomes` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'outcome_id',
  `ancestor_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'ancestor_id',
  `shortcode` char(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'ancestor_id',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'outcome_name',
  `program` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'program',
  `subject` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'subject',
  `developmental` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'developmental',
  `skills` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'skills',
  `age` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'age',
  `grade` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'grade',
  `keywords` text COLLATE utf8mb4_unicode_ci COMMENT 'keywords',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT 'description',
  `estimated_time` int(11) NOT NULL COMMENT 'estimated_time',
  `author_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'author_id',
  `author_name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'author_name',
  `organization_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'organization_id',
  `publish_scope` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'publish_scope, default as the organization_id',
  `publish_status` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'publish_status',
  `reject_reason` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'reject_reason',
  `version` int(11) NOT NULL DEFAULT '0' COMMENT 'version',
  `assumed` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'assumed',
  `locked_by` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'locked by who',
  `source_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'source_id',
  `latest_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'latest_id',
  `create_at` bigint(20) NOT NULL COMMENT 'created_at',
  `update_at` bigint(20) NOT NULL COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT NULL COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  KEY `index_ancestor_id` (`ancestor_id`),
  KEY `index_latest_id` (`latest_id`),
  KEY `index_publish_status` (`publish_status`),
  KEY `index_source_id` (`source_id`),
  FULLTEXT KEY `fullindex_name_description_keywords_shortcode` (`name`,`keywords`,`description`,`shortcode`),
  FULLTEXT KEY `fullindex_keywords` (`keywords`),
  FULLTEXT KEY `fullindex_description` (`description`),
  FULLTEXT KEY `fullindex_shortcode` (`shortcode`),
  FULLTEXT KEY `fullindex_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='outcomes table';


CREATE TABLE IF NOT EXISTS `lesson_types` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'name',
  `number` int(11) DEFAULT '0' COMMENT 'number',
  `create_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `update_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `delete_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `create_at` bigint(20) DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint(20) DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT '0' COMMENT 'delete_at',
  PRIMARY KEY (`id`),
  KEY `idx_id_delete` (`id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='lesson_types';



CREATE TABLE IF NOT EXISTS `milestones` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'name',
  `shortcode` char(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'shortcode',
  `organization_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'org id',
  `author_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'author id',
  `status` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'status',
  `describe` text COLLATE utf8mb4_unicode_ci COMMENT 'description',
  `ancestor_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'ancestor',
  `locked_by` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'who is editing',
  `source_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'previous version',
  `latest_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'latest version',
  `create_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT NULL COMMENT 'deleted_at',
  `type` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT 'normal' COMMENT 'milestone type',
  PRIMARY KEY (`id`),
  FULLTEXT KEY `fullindex_name_shortcode_describe` (`name`,`shortcode`,`describe`),
  FULLTEXT KEY `fullindex_name` (`name`),
  FULLTEXT KEY `fullindex_shortcode` (`shortcode`),
  FULLTEXT KEY `fullindex_describe` (`describe`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='milestones';


CREATE TABLE IF NOT EXISTS `milestones_outcomes` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'id',
  `milestone_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'milestone',
  `outcome_ancestor` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'outcome ancestor',
  `create_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT NULL COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  UNIQUE KEY `milestone_ancestor_id_delete` (`milestone_id`,`outcome_ancestor`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='milestones_outcomes';


CREATE TABLE IF NOT EXISTS `milestones_relations` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'id',
  `master_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'master resource',
  `relation_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'relation resource',
  `relation_type` varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'relation type',
  `master_type` varchar(30) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'master type',
  `create_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT NULL COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  UNIQUE KEY `master_relation_delete` (`master_id`,`relation_id`,`relation_type`,`master_type`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='milestones_relations';


CREATE TABLE IF NOT EXISTS `organizations_properties` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'org_id',
  `type` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'type',
  `created_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `updated_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `deleted_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `created_at` bigint(20) DEFAULT '0' COMMENT 'created_at',
  `updated_at` bigint(20) DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT '0' COMMENT 'delete_at',
  `region` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'region',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='organizations_properties';


CREATE TABLE IF NOT EXISTS `organizations_regions` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `headquarter` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '0' COMMENT 'headquarter',
  `organization_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '0' COMMENT 'organization_id',
  `create_at` bigint(20) DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint(20) DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT '0' COMMENT 'delete_at',
  PRIMARY KEY (`id`),
  KEY `organization_regions_headquarter_index` (`headquarter`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='organization_regions';


CREATE TABLE IF NOT EXISTS `outcomes_attendances` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `assessment_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'assessment id',
  `outcome_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'outcome id',
  `attendance_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'attendance id',
  PRIMARY KEY (`id`),
  KEY `outcomes_attendances_assessment_id` (`outcome_id`),
  KEY `outcomes_attendances_outcome_id` (`outcome_id`),
  KEY `outcomes_attendances_attendance_id` (`attendance_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='outcome and attendance map';


CREATE TABLE IF NOT EXISTS `outcomes_relations` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'id',
  `master_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'master resource',
  `relation_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'relation resource',
  `relation_type` varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'relation type',
  `master_type` varchar(30) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'master type',
  `create_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT NULL COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  UNIQUE KEY `master_relation_delete` (`master_id`,`relation_id`,`relation_type`,`master_type`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='outcomes_relations';


CREATE TABLE IF NOT EXISTS `outcomes_sets` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'id',
  `outcome_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'outcome_id',
  `set_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'set_id',
  `create_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT NULL COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  UNIQUE KEY `outcome_set_id_delete` (`outcome_id`,`set_id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='outcomes_sets';


CREATE TABLE IF NOT EXISTS `programs` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'name',
  `number` int(11) DEFAULT '0' COMMENT 'number',
  `org_type` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'org_type',
  `group_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'group_name',
  `create_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `update_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `delete_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `create_at` bigint(20) DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint(20) DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT '0' COMMENT 'delete_at',
  PRIMARY KEY (`id`),
  KEY `idx_id_delete` (`id`,`delete_at`),
  KEY `idx_org_type_delete_at` (`org_type`,`delete_at`),
  KEY `idx_group_name_delete_at` (`group_name`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='programs';


CREATE TABLE IF NOT EXISTS `programs_ages` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `program_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'program_id',
  `age_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'age_id',
  PRIMARY KEY (`id`),
  KEY `idx_program_id` (`program_id`),
  KEY `idx_age_id` (`age_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='programs_ages';


CREATE TABLE IF NOT EXISTS `programs_developments` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `program_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'program_id',
  `development_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'development_id',
  PRIMARY KEY (`id`),
  KEY `idx_program_id` (`program_id`),
  KEY `idx_development_id` (`development_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='programs_developments';


CREATE TABLE IF NOT EXISTS `programs_grades` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `program_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'program_id',
  `grade_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'grade_id',
  PRIMARY KEY (`id`),
  KEY `idx_program_id` (`program_id`),
  KEY `idx_grade_id` (`grade_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='programs_grades';


CREATE TABLE IF NOT EXISTS `programs_groups` (
  `program_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'program id',
  `group_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'group name',
  PRIMARY KEY (`program_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='programs groups';


CREATE TABLE IF NOT EXISTS `programs_subjects` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `program_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'program_id',
  `subject_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'subject_id',
  PRIMARY KEY (`id`),
  KEY `idx_program_id` (`program_id`),
  KEY `idx_subject_id` (`subject_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='programs_subjects';


CREATE TABLE IF NOT EXISTS `schedules` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `title` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'title',
  `class_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'class_id',
  `lesson_plan_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'lesson_plan_id',
  `org_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'org_id',
  `subject_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'subject_id',
  `program_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'program_id',
  `class_type` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'class_type',
  `start_at` bigint(20) NOT NULL COMMENT 'start_at',
  `end_at` bigint(20) NOT NULL COMMENT 'end_at',
  `due_at` bigint(20) DEFAULT NULL COMMENT 'due_at',
  `status` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'status',
  `is_all_day` tinyint(1) DEFAULT '0' COMMENT 'is_all_day',
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'description',
  `attachment` text COLLATE utf8mb4_unicode_ci COMMENT 'attachment',
  `version` bigint(20) DEFAULT '0' COMMENT 'version',
  `repeat_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'repeat_id',
  `repeat` json DEFAULT NULL COMMENT 'repeat',
  `created_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `updated_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `deleted_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `created_at` bigint(20) DEFAULT '0' COMMENT 'created_at',
  `updated_at` bigint(20) DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT '0' COMMENT 'delete_at',
  `is_hidden` tinyint(1) DEFAULT '0' COMMENT 'is hidden',
  `is_home_fun` tinyint(1) DEFAULT '0' COMMENT 'is home fun',
  PRIMARY KEY (`id`),
  KEY `schedules_org_id` (`org_id`),
  KEY `schedules_start_at` (`start_at`),
  KEY `schedules_end_at` (`end_at`),
  KEY `schedules_deleted_at` (`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='schedules';


CREATE TABLE IF NOT EXISTS `schedules_feedbacks` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `schedule_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'schedule_id',
  `user_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'user_id',
  `comment` text COLLATE utf8mb4_unicode_ci COMMENT 'Comment',
  `create_at` bigint(20) DEFAULT '0' COMMENT 'create_at',
  `update_at` bigint(20) DEFAULT '0' COMMENT 'update_at',
  `delete_at` bigint(20) DEFAULT '0' COMMENT 'delete_at',
  PRIMARY KEY (`id`),
  KEY `idx_schedule_id` (`schedule_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='schedules_feedbacks';


CREATE TABLE IF NOT EXISTS `schedules_relations` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `schedule_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'schedule_id',
  `relation_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'relation_id',
  `relation_type` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'relation_type',
  PRIMARY KEY (`id`),
  KEY `idx_schedule_id` (`schedule_id`),
  KEY `idx_relation_id` (`relation_id`),
  KEY `idx_schedule_id_relation_type` (`schedule_id`,`relation_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='schedules_relations';


CREATE TABLE IF NOT EXISTS `sets` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'name',
  `organization_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'organization_id',
  `create_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT NULL COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  KEY `index_name` (`name`),
  FULLTEXT KEY `fullindex_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='sets';


CREATE TABLE IF NOT EXISTS `skills` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'name',
  `number` int(11) DEFAULT '0' COMMENT 'number',
  `create_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `update_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `delete_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `create_at` bigint(20) DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint(20) DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT '0' COMMENT 'delete_at',
  PRIMARY KEY (`id`),
  KEY `idx_id_delete` (`id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='skills';


CREATE TABLE IF NOT EXISTS `subjects` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'name',
  `number` int(11) DEFAULT '0' COMMENT 'number',
  `create_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `update_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `delete_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `create_at` bigint(20) DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint(20) DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT '0' COMMENT 'delete_at',
  PRIMARY KEY (`id`),
  KEY `idx_id_delete` (`id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='subjects';


CREATE TABLE IF NOT EXISTS `user_settings` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `user_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'user_id',
  `setting_json` json DEFAULT NULL COMMENT 'setting_json',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='user_settings';


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


CREATE TABLE IF NOT EXISTS `visibility_settings` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'name',
  `number` int(11) DEFAULT '0' COMMENT 'number',
  `create_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `update_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `delete_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `create_at` bigint(20) DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint(20) DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT '0' COMMENT 'delete_at',
  PRIMARY KEY (`id`),
  KEY `idx_id_delete` (`id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='visibility_settings';


INSERT IGNORE INTO `class_types` (`id`, `name`, `create_id`, `update_id`, `delete_id`, `create_at`, `update_at`, `delete_at`, `number`) VALUES ('Homework','schedule_detail_homework',NULL,NULL,NULL,0,0,0,0),('OfflineClass','schedule_detail_offline_class',NULL,NULL,NULL,0,0,0,0),('OnlineClass','schedule_detail_online_class',NULL,NULL,NULL,0,0,0,0),('Task','schedule_detail_task',NULL,NULL,NULL,0,0,0,0);

INSERT IGNORE INTO `lesson_types` (`id`, `name`, `create_id`, `update_id`, `delete_id`, `create_at`, `update_at`, `delete_at`, `number`) VALUES ('1','library_label_test',NULL,NULL,NULL,0,0,0,0),('2','library_label_not_test',NULL,NULL,NULL,0,0,0,0);

INSERT IGNORE INTO `organizations_properties` (`id`, `type`, `created_id`, `updated_id`, `deleted_id`, `created_at`, `updated_at`, `delete_at`, `region`) VALUES ('10f38ce9-5152-4049-b4e7-6d2e2ba884e6','headquarters',NULL,NULL,NULL,0,0,0,'global'),('9d42af2a-d943-4bb7-84d8-9e2e28b0e290','headquarters',NULL,NULL,NULL,0,0,0,'vn');

INSERT IGNORE INTO `organizations_regions` (`id`, `headquarter`, `organization_id`, `create_at`, `update_at`, `delete_at`) VALUES ('5fb24528993e7591084c2c46','9d42af2a-d943-4bb7-84d8-9e2e28b0e290','281e49c6-a1f8-4d5e-83f2-0cf76700601c',1615963415,1615963415,0),('5fb24528993e7591084c2c47','9d42af2a-d943-4bb7-84d8-9e2e28b0e290','f7d55488-a419-4436-b6d6-5d9021be388c',1618474165,1618474165,0),('5fb24528993e7591084c2c48','9d42af2a-d943-4bb7-84d8-9e2e28b0e290','69ddc951-f20b-4792-9b66-455f371491e9',1620640679,1620640679,0),('5fb24528993e7591084c2c49','9d42af2a-d943-4bb7-84d8-9e2e28b0e290','20cb7b33-35c2-4074-95a3-c782dc4fc1fd',1620640679,1620640679,0),('5fb24528993e7591084c2c50','9d42af2a-d943-4bb7-84d8-9e2e28b0e290','6716f8ec-5470-4d5f-b3e2-d3af043595e6',1620640679,1620640679,0),('5fb24528993e7591084c2c51','9d42af2a-d943-4bb7-84d8-9e2e28b0e290','69ba14b8-7198-4b12-8566-349e0767bc50',1620640679,1620640679,0),('5fb24528993e7591084c2c52','9d42af2a-d943-4bb7-84d8-9e2e28b0e290','7aa8874d-7bea-4915-b832-de2d8506741c',1620640679,1620640679,0),('5fb24528993e7591084c2c53','9d42af2a-d943-4bb7-84d8-9e2e28b0e290','53ddf7ac-e87f-4048-a641-6b1e1dc7b484',1620640679,1620640679,0),('5fb24528993e7591084c2c55','9d42af2a-d943-4bb7-84d8-9e2e28b0e290','ab54b78e-7f6f-4464-8022-27413d1af20f',1620640679,1620640679,0),('5fb24528993e7591084c2c56','9d42af2a-d943-4bb7-84d8-9e2e28b0e290','c37b7446-1807-4c31-bcb5-90d23d1c808a',1620640679,1620640679,0);

INSERT IGNORE INTO `programs_groups` (`program_id`, `group_name`) VALUES ('4591423a-2619-4ef8-a900-f5d924939d02','BadaSTEAM'),('56e24fa0-e139-4c80-b365-61c9bc42cd3f','BadaESL'),('7565ae11-8130-4b7d-ac24-1d9dd6f792f2','More'),('7a8c5021-142b-44b1-b60b-275c29d132fe','BadaESL'),('93f293e8-2c6a-47ad-bc46-1554caac99e4','BadaESL'),('b39edb9a-ab91-4245-94a4-eb2b5007c033','BadaESL'),('d1bbdcc5-0d80-46b0-b98e-162e7439058f','BadaSTEAM'),('f6617737-5022-478d-9672-0354667e0338','BadaESL');

INSERT IGNORE INTO `user_settings` (`id`, `user_id`, `setting_json`) VALUES ('default_setting_0','default_setting_0','{\"cms_page_size\": 20}');

INSERT IGNORE INTO `visibility_settings` (`id`, `name`, `create_id`, `update_id`, `delete_id`, `create_at`, `update_at`, `delete_at`, `number`) VALUES ('visibility_settings1','library_label_visibility_school',NULL,NULL,NULL,0,0,0,0),('visibility_settings2','library_label_visibility_organization',NULL,NULL,NULL,0,0,0,0);
