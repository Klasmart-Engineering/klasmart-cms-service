-- MySQL dump 10.13  Distrib 5.7.32, for Linux (x86_64)
--
-- Host: 127.0.0.1    Database: kidsloop2
-- ------------------------------------------------------
-- Server version	8.0.21

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `ages`
--

DROP TABLE IF EXISTS `ages`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ages` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'name',
  `create_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `update_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `delete_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `create_at` bigint DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint DEFAULT '0' COMMENT 'delete_at',
  `number` int DEFAULT '0' COMMENT 'number',
  PRIMARY KEY (`id`),
  KEY `idx_id_delete` (`id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='ages';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessments`
--

DROP TABLE IF EXISTS `assessments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `assessments` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `schedule_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'schedule id',
  `title` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'title',
  `program_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'program id',
  `subject_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'subject id',
  `teacher_ids` json NOT NULL COMMENT 'teacher id',
  `class_length` int NOT NULL COMMENT 'class length (util: minute)',
  `class_end_time` bigint NOT NULL COMMENT 'class end time (unix seconds)',
  `complete_time` bigint NOT NULL COMMENT 'complete time (unix seconds)',
  `status` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'status (enum: in_progress, complete)',
  `create_at` bigint NOT NULL COMMENT 'create time (unix seconds)',
  `update_at` bigint NOT NULL COMMENT 'update time (unix seconds)',
  `delete_at` bigint NOT NULL COMMENT 'delete time (unix seconds)',
  PRIMARY KEY (`id`),
  KEY `assessments_status` (`status`),
  KEY `assessments_schedule_id` (`schedule_id`),
  KEY `assessments_complete_time` (`complete_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='assessment';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessments_attendances`
--

DROP TABLE IF EXISTS `assessments_attendances`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `assessments_attendances` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `assessment_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'assessment id',
  `attendance_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'attendance id',
  `checked` tinyint(1) NOT NULL COMMENT 'checked',
  PRIMARY KEY (`id`),
  KEY `assessments_attendances_assessment_id` (`assessment_id`),
  KEY `assessments_attendances_attendance_id` (`attendance_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='assessment and attendance map';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `assessments_outcomes`
--

DROP TABLE IF EXISTS `assessments_outcomes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `assessments_outcomes` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `assessment_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'assessment id',
  `outcome_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'outcome id',
  `skip` tinyint(1) NOT NULL COMMENT 'skip',
  `none_achieved` tinyint(1) NOT NULL COMMENT 'none achieved',
  PRIMARY KEY (`id`),
  KEY `assessments_outcomes_assessment_id` (`assessment_id`),
  KEY `assessments_outcomes_outcome_id` (`outcome_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='assessment and outcome map';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `class_types`
--

DROP TABLE IF EXISTS `class_types`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `class_types` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'name',
  `create_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `update_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `delete_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `create_at` bigint DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint DEFAULT '0' COMMENT 'delete_at',
  `number` int DEFAULT '0' COMMENT 'number',
  PRIMARY KEY (`id`),
  KEY `idx_id_delete` (`id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='class_types';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cms_authed_contents`
--

DROP TABLE IF EXISTS `cms_authed_contents`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cms_authed_contents` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'record_id',
  `org_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'org_id',
  `from_folder_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'from_folder_id',
  `content_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'content_id',
  `creator` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'creator',
  `duration` int NOT NULL DEFAULT '0' COMMENT 'duration',
  `create_at` bigint NOT NULL DEFAULT '0' COMMENT 'created_at',
  `delete_at` bigint DEFAULT '0' COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  KEY `org_id` (`org_id`),
  KEY `content_id` (`content_id`),
  KEY `creator` (`creator`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='内容授权记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cms_contents`
--

DROP TABLE IF EXISTS `cms_contents`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cms_contents` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'content_id',
  `content_type` int NOT NULL COMMENT '数据类型',
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
  `suggest_time` int NOT NULL COMMENT '建议时间',
  `author` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '作者id',
  `creator` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '创建者id',
  `org` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '所属机构',
  `publish_scope` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '发布范围',
  `publish_status` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '状态',
  `reject_reason` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '拒绝理由',
  `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '拒绝理由备注',
  `version` int NOT NULL DEFAULT '0' COMMENT '版本',
  `locked_by` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '封锁人',
  `source_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'source_id',
  `latest_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'latest_id',
  `lesson_type` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'lesson_type',
  `create_at` bigint NOT NULL COMMENT 'created_at',
  `update_at` bigint NOT NULL COMMENT 'updated_at',
  `delete_at` bigint DEFAULT NULL COMMENT 'deleted_at',
  `self_study` tinyint NOT NULL DEFAULT '0' COMMENT '是否支持自学',
  `draw_activity` tinyint NOT NULL DEFAULT '0' COMMENT '是否支持绘画',
  `copy_source_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'copy_source_id',
  PRIMARY KEY (`id`),
  KEY `content_type` (`content_type`),
  KEY `content_author` (`author`),
  KEY `content_org` (`org`),
  KEY `content_publish_status` (`publish_status`),
  KEY `content_source_id` (`source_id`),
  KEY `content_latest_id` (`latest_id`),
  FULLTEXT KEY `content_name_description_keywords_author_index` (`content_name`,`keywords`,`description`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='内容表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cms_folder_items`
--

DROP TABLE IF EXISTS `cms_folder_items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cms_folder_items` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `owner_type` int NOT NULL COMMENT 'folder item owner type',
  `owner` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'folder item owner',
  `parent_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'folder item parent folder id',
  `link` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'folder item link',
  `item_type` int NOT NULL COMMENT 'folder item type',
  `dir_path` varchar(2048) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'folder item path',
  `editor` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'folder item editor',
  `items_count` int NOT NULL COMMENT 'folder item count',
  `name` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'folder item name',
  `partition` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'folder item partition',
  `thumbnail` text COLLATE utf8mb4_unicode_ci COMMENT 'folder item thumbnail',
  `creator` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'folder item creator',
  `create_at` bigint NOT NULL COMMENT 'create time (unix seconds)',
  `update_at` bigint NOT NULL COMMENT 'update time (unix seconds)',
  `delete_at` bigint DEFAULT NULL COMMENT 'delete time (unix seconds)',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='cms folder';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cms_shared_folders`
--

DROP TABLE IF EXISTS `cms_shared_folders`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cms_shared_folders` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'record_id',
  `folder_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'folder_id',
  `org_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'org_id',
  `creator` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'creator',
  `create_at` bigint NOT NULL DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint NOT NULL DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint DEFAULT '0' COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  KEY `org_id` (`org_id`),
  KEY `folder_id` (`folder_id`),
  KEY `creator` (`creator`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件夹分享记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `developmentals`
--

DROP TABLE IF EXISTS `developmentals`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `developmentals` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'name',
  `create_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `update_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `delete_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `create_at` bigint DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint DEFAULT '0' COMMENT 'delete_at',
  `number` int DEFAULT '0' COMMENT 'number',
  PRIMARY KEY (`id`),
  KEY `idx_id_delete` (`id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='developmentals';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `developments_skills`
--

DROP TABLE IF EXISTS `developments_skills`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `developments_skills` (
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `grades`
--

DROP TABLE IF EXISTS `grades`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `grades` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'name',
  `create_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `update_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `delete_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `create_at` bigint DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint DEFAULT '0' COMMENT 'delete_at',
  `number` int DEFAULT '0' COMMENT 'number',
  PRIMARY KEY (`id`),
  KEY `idx_id_delete` (`id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='grades';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `learning_outcomes`
--

DROP TABLE IF EXISTS `learning_outcomes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
  `estimated_time` int NOT NULL COMMENT 'estimated_time',
  `author_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'author_id',
  `author_name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'author_name',
  `organization_id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'organization_id',
  `publish_scope` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'publish_scope, default as the organization_id',
  `publish_status` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'publish_status',
  `reject_reason` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'reject_reason',
  `version` int NOT NULL DEFAULT '0' COMMENT 'version',
  `assumed` tinyint NOT NULL DEFAULT '0' COMMENT 'assumed',
  `locked_by` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'locked by who',
  `source_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'source_id',
  `latest_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'latest_id',
  `create_at` bigint NOT NULL COMMENT 'created_at',
  `update_at` bigint NOT NULL COMMENT 'updated_at',
  `delete_at` bigint DEFAULT NULL COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  KEY `index_ancestor_id` (`ancestor_id`),
  KEY `index_latest_id` (`latest_id`),
  KEY `index_publish_status` (`publish_status`),
  KEY `index_source_id` (`source_id`),
  FULLTEXT KEY `fullindex_name_description_keywords_shortcode` (`name`,`keywords`,`description`,`shortcode`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='outcomes table';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `lesson_types`
--

DROP TABLE IF EXISTS `lesson_types`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `lesson_types` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'name',
  `create_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `update_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `delete_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `create_at` bigint DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint DEFAULT '0' COMMENT 'delete_at',
  `number` int DEFAULT '0' COMMENT 'number',
  PRIMARY KEY (`id`),
  KEY `idx_id_delete` (`id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='lesson_types';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `organizations_properties`
--

DROP TABLE IF EXISTS `organizations_properties`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `organizations_properties` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'org_id',
  `type` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'type',
  `created_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `updated_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `deleted_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `created_at` bigint DEFAULT '0' COMMENT 'created_at',
  `updated_at` bigint DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint DEFAULT '0' COMMENT 'delete_at',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='organizations_properties';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `outcomes_attendances`
--

DROP TABLE IF EXISTS `outcomes_attendances`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `programs`
--

DROP TABLE IF EXISTS `programs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `programs` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'name',
  `create_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `update_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `delete_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `create_at` bigint DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint DEFAULT '0' COMMENT 'delete_at',
  `number` int DEFAULT '0' COMMENT 'number',
  `org_type` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'org_type',
  `group_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'group_name',
  PRIMARY KEY (`id`),
  KEY `idx_id_delete` (`id`,`delete_at`),
  KEY `idx_org_type_delete_at` (`org_type`,`delete_at`),
  KEY `idx_group_name_delete_at` (`group_name`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='programs';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `programs_ages`
--

DROP TABLE IF EXISTS `programs_ages`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `programs_ages` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `program_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'program_id',
  `age_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'age_id',
  PRIMARY KEY (`id`),
  KEY `idx_program_id` (`program_id`),
  KEY `idx_age_id` (`age_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='programs_ages';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `programs_developments`
--

DROP TABLE IF EXISTS `programs_developments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `programs_developments` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `program_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'program_id',
  `development_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'development_id',
  PRIMARY KEY (`id`),
  KEY `idx_program_id` (`program_id`),
  KEY `idx_development_id` (`development_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='programs_developments';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `programs_grades`
--

DROP TABLE IF EXISTS `programs_grades`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `programs_grades` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `program_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'program_id',
  `grade_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'grade_id',
  PRIMARY KEY (`id`),
  KEY `idx_program_id` (`program_id`),
  KEY `idx_grade_id` (`grade_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='programs_grades';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `programs_subjects`
--

DROP TABLE IF EXISTS `programs_subjects`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `programs_subjects` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `program_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'program_id',
  `subject_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'subject_id',
  PRIMARY KEY (`id`),
  KEY `idx_program_id` (`program_id`),
  KEY `idx_subject_id` (`subject_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='programs_subjects';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `schedules`
--

DROP TABLE IF EXISTS `schedules`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `schedules` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `title` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'title',
  `class_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'class_id',
  `lesson_plan_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'lesson_plan_id',
  `org_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'org_id',
  `subject_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'subject_id',
  `program_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'program_id',
  `class_type` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'class_type',
  `start_at` bigint NOT NULL COMMENT 'start_at',
  `end_at` bigint NOT NULL COMMENT 'end_at',
  `due_at` bigint DEFAULT NULL COMMENT 'due_at',
  `status` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'status',
  `is_all_day` tinyint(1) DEFAULT '0' COMMENT 'is_all_day',
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'description',
  `attachment` text COLLATE utf8mb4_unicode_ci COMMENT 'attachment',
  `version` bigint DEFAULT '0' COMMENT 'version',
  `repeat_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'repeat_id',
  `repeat` json DEFAULT NULL COMMENT 'repeat',
  `created_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `updated_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `deleted_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `created_at` bigint DEFAULT '0' COMMENT 'created_at',
  `updated_at` bigint DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint DEFAULT '0' COMMENT 'delete_at',
  PRIMARY KEY (`id`),
  KEY `schedules_org_id` (`org_id`),
  KEY `schedules_start_at` (`start_at`),
  KEY `schedules_end_at` (`end_at`),
  KEY `schedules_deleted_at` (`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='schedules';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `skills`
--

DROP TABLE IF EXISTS `skills`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `skills` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'name',
  `create_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `update_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `delete_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `create_at` bigint DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint DEFAULT '0' COMMENT 'delete_at',
  `developmental_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'development_id',
  `number` int DEFAULT '0' COMMENT 'number',
  PRIMARY KEY (`id`),
  KEY `idx_id_delete` (`id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='skills';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `subjects`
--

DROP TABLE IF EXISTS `subjects`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `subjects` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'name',
  `create_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `update_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `delete_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `create_at` bigint DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint DEFAULT '0' COMMENT 'delete_at',
  `number` int DEFAULT '0' COMMENT 'number',
  PRIMARY KEY (`id`),
  KEY `idx_id_delete` (`id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='subjects';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_settings`
--

DROP TABLE IF EXISTS `user_settings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `user_settings` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `user_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'user_id',
  `setting_json` json DEFAULT NULL COMMENT 'setting_json',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='user_settings';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `visibility_settings`
--

DROP TABLE IF EXISTS `visibility_settings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `visibility_settings` (
  `id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'name',
  `create_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `update_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `delete_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
  `create_at` bigint DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint DEFAULT '0' COMMENT 'delete_at',
  `number` int DEFAULT '0' COMMENT 'number',
  PRIMARY KEY (`id`),
  KEY `idx_id_delete` (`id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='visibility_settings';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2020-12-30  7:57:57
