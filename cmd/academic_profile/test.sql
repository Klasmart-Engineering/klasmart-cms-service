drop database if exists kidsloop2_mapping;

create database kidsloop2_mapping;
use kidsloop2_mapping;

CREATE TABLE `assessments` (
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

insert into assessments 
select `id`, `schedule_id`, `title`, `program_id`, `subject_id`, `teacher_ids`, `class_length`, `class_end_time`, `complete_time`, `status`, `create_at`, `update_at`, `delete_at`, `type`
from kidsloop_alpha.assessments;

CREATE TABLE `assessments_attendances` (
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

insert into assessments_attendances
select `id`, `assessment_id`, `attendance_id`, `checked`, `origin`, `role`
from kidsloop_alpha.assessments_attendances;

CREATE TABLE `assessments_contents` (
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

insert into assessments_contents
select `id`, `assessment_id`, `content_id`, `content_name`, `content_type`, `content_comment`, `checked` 
from kidsloop_alpha.assessments_contents;

CREATE TABLE `assessments_contents_outcomes` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'id',
  `assessment_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'assessment id',
  `content_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'content id',
  `outcome_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'outcome id',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_assessments_contents_outcomes` (`assessment_id`,`content_id`,`outcome_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='assessment content and outcome map';

insert into assessments_contents_outcomes
select `id`, `assessment_id`, `content_id`, `outcome_id`
from kidsloop_alpha.assessments_contents_outcomes;

CREATE TABLE `assessments_outcomes` (
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

insert into assessments_outcomes
select `id`, `assessment_id`, `outcome_id`, `skip`, `none_achieved`, `checked`
from kidsloop_alpha.assessments_outcomes;

CREATE TABLE `cms_authed_contents` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'record_id',
  `org_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'org_id',
  `from_folder_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'from_folder_id',
  `content_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'content_id',
  `creator` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'creator',
  `duration` int(11) NOT NULL DEFAULT '0' COMMENT 'duration',
  `create_at` bigint(20) NOT NULL COMMENT 'created_at',
  `delete_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  KEY `org_id` (`org_id`),
  KEY `content_id` (`content_id`),
  KEY `creator` (`creator`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='内容授权记录表';

insert into cms_authed_contents
select `id`, `org_id`, `from_folder_id`, `content_id`, `creator`, `duration`, `create_at`, `delete_at`
from kidsloop_alpha.cms_authed_contents;

CREATE TABLE `cms_content_properties` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `content_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'content id',
  `property_type` int(11) NOT NULL DEFAULT '0' COMMENT 'property type',
  `property_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'property id',
  `sequence` int(11) NOT NULL DEFAULT '0' COMMENT 'sequence',
  PRIMARY KEY (`id`),
  KEY `cms_content_properties_content_id_idx` (`content_id`),
  KEY `cms_content_properties_property_type_idx` (`property_type`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='cms content properties';

insert into cms_content_properties
select `id`, `content_id`, `property_type`, `property_id`, `sequence`
from kidsloop_alpha.cms_content_properties;

CREATE TABLE `cms_content_visibility_settings` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `content_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'content id',
  `visibility_setting` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'visibility setting',
  PRIMARY KEY (`id`),
  KEY `cms_content_visibility_settings_content_id_idx` (`content_id`),
  KEY `cms_content_visibility_settings_visibility_settings_idx` (`visibility_setting`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='cms content visibility settings';

insert into cms_content_visibility_settings
select `id`, `content_id`, `visibility_setting`
from kidsloop_alpha.cms_content_visibility_settings;

CREATE TABLE `cms_contents` (
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
  `copy_source_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'copy_source_id',
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

insert into cms_contents
select `id`, `content_type`, `content_name`, `program`, `subject`, `developmental`, `skills`, `age`, `grade`, `keywords`, `description`, `thumbnail`, `source_type`, `data`, `extra`, `outcomes`, `dir_path`, `suggest_time`, `author`, `creator`, `org`, `publish_scope`, `publish_status`, `self_study`, `draw_activity`, `reject_reason`, `remark`, `version`, `locked_by`, `source_id`, `copy_source_id`, `latest_id`, `lesson_type`, `create_at`, `update_at`, `delete_at`
from kidsloop_alpha.cms_contents;

CREATE TABLE `cms_folder_items` (
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
  `keywords` text COLLATE utf8mb4_unicode_ci COMMENT '???',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '??',
  PRIMARY KEY (`id`),
  FULLTEXT KEY `folder_name_description_keywords_author_index` (`name`,`keywords`,`description`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='cms folder';

insert into cms_folder_items
select `id`, `owner_type`, `owner`, `parent_id`, `link`, `item_type`, `dir_path`, `editor`, `items_count`, `name`, `partition`, `thumbnail`, `creator`, `create_at`, `update_at`, `delete_at`, `keywords`, `description` 
from kidsloop_alpha.cms_folder_items;

CREATE TABLE `cms_shared_folders` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'record_id',
  `folder_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'folder_id',
  `org_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'org_id',
  `creator` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'creator',
  `create_at` bigint(20) NOT NULL COMMENT 'created_at',
  `update_at` bigint(20) NOT NULL COMMENT 'updated_at',
  `delete_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  KEY `org_id` (`org_id`),
  KEY `folder_id` (`folder_id`),
  KEY `creator` (`creator`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件夹分享记录表';

insert into cms_shared_folders
select `id`, `folder_id`, `org_id`, `creator`, `create_at`, `update_at`, `delete_at`
from kidsloop_alpha.cms_shared_folders;

CREATE TABLE `feedbacks_assignments` (
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

insert into feedbacks_assignments
select `id`, `feedback_id`, `attachment_id`, `attachment_name`, `number`, `create_at`, `update_at`, `delete_at`
from kidsloop_alpha.feedbacks_assignments;

CREATE TABLE `home_fun_studies` (
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

insert into home_fun_studies
select `id`, `schedule_id`, `title`, `teacher_ids`, `student_id`, `subject_id`, `status`, `due_at`, `complete_at`, `latest_feedback_id`, `latest_feedback_at`, `assess_feedback_id`, `assess_score`, `assess_comment`, `create_at`, `update_at`, `delete_at`
from kidsloop_alpha.home_fun_studies;

CREATE TABLE `learning_outcomes` (
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

insert into learning_outcomes
select `id`, `ancestor_id`, `shortcode`, `name`, `program`, `subject`, `developmental`, `skills`, `age`, `grade`, `keywords`, `description`, `estimated_time`, `author_id`, `author_name`, `organization_id`, `publish_scope`, `publish_status`, `reject_reason`, `version`, `assumed`, `locked_by`, `source_id`, `latest_id`, `create_at`, `update_at`, `delete_at`
from kidsloop_alpha.learning_outcomes;

CREATE TABLE `milestones` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT '',
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

insert into milestones
select `id`, `name`, `shortcode`, `organization_id`, `author_id`, `status`, `describe`, `ancestor_id`, `locked_by`, `source_id`, `latest_id`, `create_at`, `update_at`, `delete_at`, `type`
from kidsloop_alpha.milestones;

CREATE TABLE `milestones_outcomes` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'id',
  `milestone_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'milestone',
  `outcome_ancestor` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'outcome ancestor',
  `create_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT NULL COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  UNIQUE KEY `milestone_ancestor_id_delete` (`milestone_id`,`outcome_ancestor`,`delete_at`)
) ENGINE=InnoDB AUTO_INCREMENT=156 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='milestones_outcomes';

insert into milestones_outcomes
select `id`, `milestone_id`, `outcome_ancestor`, `create_at`, `update_at`, `delete_at`
from kidsloop_alpha.milestones_outcomes;

CREATE TABLE `milestones_relations` (
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
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='milestones_relations';

insert into milestones_relations
select `id`, `master_id`, `relation_id`, `relation_type`, `master_type`, `create_at`, `update_at`, `delete_at`
from kidsloop_alpha.milestones_relations;

CREATE TABLE `outcomes_attendances` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `assessment_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'assessment id',
  `outcome_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'outcome id',
  `attendance_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'attendance id',
  PRIMARY KEY (`id`),
  KEY `outcomes_attendances_assessment_id` (`outcome_id`),
  KEY `outcomes_attendances_outcome_id` (`outcome_id`),
  KEY `outcomes_attendances_attendance_id` (`attendance_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='outcome and attendance map';

insert into outcomes_attendances
select `id`, `assessment_id`, `outcome_id`, `attendance_id`
from kidsloop_alpha.outcomes_attendances;

CREATE TABLE `outcomes_relations` (
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
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='outcomes_relations';

insert into outcomes_relations
select `id`, `master_id`, `relation_id`, `relation_type`, `master_type`, `create_at`, `update_at`, `delete_at`
from kidsloop_alpha.outcomes_relations;

CREATE TABLE `outcomes_sets` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'id',
  `outcome_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'outcome_id',
  `set_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'set_id',
  `create_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT NULL COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  UNIQUE KEY `outcome_set_id_delete` (`outcome_id`,`set_id`,`delete_at`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='outcomes_sets';

insert into outcomes_sets
select `id`, `outcome_id`, `set_id`, `create_at`, `update_at`, `delete_at`
from kidsloop_alpha.outcomes_sets;

CREATE TABLE `schedules` (
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

insert into schedules
select `id`, `title`, `class_id`, `lesson_plan_id`, `org_id`, `subject_id`, `program_id`, `class_type`, `start_at`, `end_at`, `due_at`, `status`, `is_all_day`, `description`, `attachment`, `version`, `repeat_id`, `repeat`, `created_id`, `updated_id`, `deleted_id`, `created_at`, `updated_at`, `delete_at`, `is_hidden`, `is_home_fun`
from kidsloop_alpha.schedules;

CREATE TABLE `schedules_relations` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `schedule_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'schedule_id',
  `relation_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'relation_id',
  `relation_type` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'record_type',
  PRIMARY KEY (`id`),
  KEY `idx_schedule_id` (`schedule_id`),
  KEY `idx_relation_id` (`relation_id`),
  KEY `idx_schedule_id_relation_type` (`schedule_id`,`relation_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='schedules_relations';

insert into schedules_relations
select `id`, `schedule_id`, `relation_id`, `relation_type`
from kidsloop_alpha.schedules_relations;

CREATE TABLE `sets` (
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

insert into `sets`
select `id`, `name`, `organization_id`, `create_at`, `update_at`, `delete_at`
from kidsloop_alpha.sets;