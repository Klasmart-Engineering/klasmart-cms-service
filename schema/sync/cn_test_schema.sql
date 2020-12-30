-- MySQL dump 10.13  Distrib 5.7.32, for Linux (x86_64)
--
-- Host: 127.0.0.1    Database: kidsloop2
-- ------------------------------------------------------
-- Server version	8.0.20

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
-- Table structure for table `assessments`
--

DROP TABLE IF EXISTS `assessments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `assessments` (
  `id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `schedule_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'schedule id',
  `title` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'title',
  `program_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'program id',
  `subject_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'subject id',
  `teacher_ids` json NOT NULL COMMENT 'teacher id',
  `class_length` int NOT NULL COMMENT 'class length (util: minute)',
  `class_end_time` bigint NOT NULL COMMENT 'class end time (unix seconds)',
  `complete_time` bigint NOT NULL COMMENT 'complete time (unix seconds)',
  `status` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'status (enum: in_progress, complete)',
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
  `id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `assessment_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'assessment id',
  `attendance_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'attendance id',
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
  `id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `assessment_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'assessment id',
  `outcome_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'outcome id',
  `skip` tinyint(1) NOT NULL COMMENT 'skip',
  PRIMARY KEY (`id`),
  KEY `assessments_outcomes_assessment_id` (`assessment_id`),
  KEY `assessments_outcomes_outcome_id` (`outcome_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='assessment and outcome map';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cms_contents`
--

DROP TABLE IF EXISTS `cms_contents`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cms_contents` (
  `id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'content_id',
  `content_type` int NOT NULL COMMENT '数据类型',
  `content_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '内容名称',
  `program` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'program',
  `subject` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'subject',
  `developmental` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'developmental',
  `skills` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'skills',
  `age` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'age',
  `grade` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'grade',
  `keywords` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '关键字',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '描述',
  `thumbnail` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '封面',
  `data` json DEFAULT NULL COMMENT '数据',
  `extra` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '附加数据',
  `outcomes` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'Learning outcomes',
  `suggest_time` int NOT NULL COMMENT '建议时间',
  `author` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '作者id',
  `author_name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '作者名',
  `org` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '所属机构',
  `publish_scope` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '发布范围',
  `publish_status` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '状态',
  `reject_reason` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '拒绝理由',
  `version` int NOT NULL DEFAULT '0' COMMENT '版本',
  `locked_by` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '封锁人',
  `source_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'source_id',
  `latest_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'latest_id',
  `create_at` bigint NOT NULL COMMENT 'created_at',
  `update_at` bigint NOT NULL COMMENT 'updated_at',
  `delete_at` bigint DEFAULT NULL COMMENT 'deleted_at',
  `source_type` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'content source type',
  `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'reject reason remark',
  `self_study` tinyint NOT NULL DEFAULT '0' COMMENT 'is content can self study',
  `draw_activity` tinyint NOT NULL DEFAULT '0' COMMENT 'is activity can draw',
  `lesson_type` tinyint NOT NULL DEFAULT '0' COMMENT 'lesson type',
  PRIMARY KEY (`id`),
  KEY `content_type` (`content_type`),
  KEY `content_author` (`author`),
  KEY `content_org` (`org`),
  KEY `content_publish_status` (`publish_status`),
  KEY `content_source_id` (`source_id`),
  KEY `content_latest_id` (`latest_id`),
  FULLTEXT KEY `content_name_index` (`content_name`) /*!50100 WITH PARSER `ngram` */ ,
  FULLTEXT KEY `content_description_index` (`keywords`) /*!50100 WITH PARSER `ngram` */ ,
  FULLTEXT KEY `content_keywords_index` (`description`) /*!50100 WITH PARSER `ngram` */ ,
  FULLTEXT KEY `content_author_index` (`author_name`) /*!50100 WITH PARSER `ngram` */ ,
  FULLTEXT KEY `content_name_description_keywords_author_index` (`content_name`,`keywords`,`description`,`author_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='内容表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `learning_outcomes`
--

DROP TABLE IF EXISTS `learning_outcomes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `learning_outcomes` (
  `id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'outcome_id',
  `ancestor_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'ancestor_id',
  `shortcode` char(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'ancestor_id',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'outcome_name',
  `program` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'program',
  `subject` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'subject',
  `developmental` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'developmental',
  `skills` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'skills',
  `age` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'age',
  `grade` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'grade',
  `keywords` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'keywords',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'description',
  `estimated_time` int NOT NULL COMMENT 'estimated_time',
  `author_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'author_id',
  `author_name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'author_name',
  `organization_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'organization_id',
  `publish_scope` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'publish_scope, default as the organization_id',
  `publish_status` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'publish_status',
  `reject_reason` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'reject_reason',
  `version` int NOT NULL DEFAULT '0' COMMENT 'version',
  `assumed` tinyint NOT NULL DEFAULT '0' COMMENT 'assumed',
  `locked_by` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'locked by who',
  `source_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'source_id',
  `latest_id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'latest_id',
  `create_at` bigint NOT NULL COMMENT 'created_at',
  `update_at` bigint NOT NULL COMMENT 'updated_at',
  `delete_at` bigint DEFAULT NULL COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  KEY `index_ancestor_id` (`ancestor_id`),
  KEY `index_latest_id` (`latest_id`),
  KEY `index_publish_status` (`publish_status`),
  KEY `index_source_id` (`source_id`),
  FULLTEXT KEY `fullindex_name_description_keywords_author_shortcode` (`name`,`keywords`,`description`,`author_name`,`shortcode`) /*!50100 WITH PARSER `ngram` */ 
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='outcomes table';
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
  `id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `assessment_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'assessment id',
  `outcome_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'outcome id',
  `attendance_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'attendance id',
  PRIMARY KEY (`id`),
  KEY `outcomes_attendances_assessment_id` (`outcome_id`),
  KEY `outcomes_attendances_outcome_id` (`outcome_id`),
  KEY `outcomes_attendances_attendance_id` (`attendance_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='outcome and attendance map';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `schedules`
--

DROP TABLE IF EXISTS `schedules`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `schedules` (
  `id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `title` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'title',
  `class_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'class_id',
  `lesson_plan_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'lesson_plan_id',
  `org_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'org_id',
  `subject_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'subject_id',
  `program_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'program_id',
  `class_type` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'class_type',
  `start_at` bigint NOT NULL COMMENT 'start_at',
  `end_at` bigint NOT NULL COMMENT 'end_at',
  `due_at` bigint DEFAULT NULL COMMENT 'due_at',
  `status` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'status',
  `is_all_day` tinyint(1) DEFAULT '0' COMMENT 'is_all_day',
  `description` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'description',
  `attachment` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'attachment',
  `version` bigint DEFAULT '0' COMMENT 'version',
  `repeat_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'repeat_id',
  `repeat` json DEFAULT NULL COMMENT 'repeat',
  `created_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'created_id',
  `updated_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'updated_id',
  `deleted_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'deleted_id',
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
-- Table structure for table `schedules_teachers`
--

DROP TABLE IF EXISTS `schedules_teachers`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `schedules_teachers` (
  `id` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `teacher_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'teacher_id',
  `schedule_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'schedule_id',
  `delete_at` bigint DEFAULT '0' COMMENT 'delete_at',
  PRIMARY KEY (`id`),
  KEY `schedules_teacher_id` (`teacher_id`),
  KEY `schedules_schedule_id` (`schedule_id`),
  KEY `schedules_deleted_at` (`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='schedules_teachers';
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2020-12-30  6:51:02
