
CREATE TABLE IF NOT EXISTS `assessments_v2` (
    `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
    `org_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'org id',
    `schedule_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'schedule id',
    `assessment_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'assessment type',
    `title` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'title',
    `status` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'status',
    `complete_at` bigint(20) DEFAULT '0' COMMENT 'complete time (unix seconds)',
    `class_length` int(11) DEFAULT '0' COMMENT 'class length (util: minute)',
    `class_end_at` bigint(20) DEFAULT '0' COMMENT 'class end time (unix seconds)',
    `migrate_flag` int(11) DEFAULT '0' COMMENT 'assisted migration',

    `create_at` bigint(20) NOT NULL COMMENT 'create time (unix seconds)',
    `update_at` bigint(20) NOT NULL COMMENT 'update time (unix seconds)',
    `delete_at` bigint(20) NOT NULL COMMENT 'delete time (unix seconds)',
    PRIMARY KEY (`id`),
    KEY `assessments_org_id` (`org_id`),
    KEY `assessments_schedule_id` (`schedule_id`),
    KEY `assessments_type` (`assessment_type`),
    KEY `assessments_status` (`status`),
    KEY `assessments_delete_at` (`delete_at`),

    UNIQUE `assessments_unique` (`schedule_id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='assessments_v2';

ALTER TABLE `assessments_v2` ADD INDEX assessments_complete_at(`complete_at`);

CREATE TABLE IF NOT EXISTS `assessments_users_v2` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `assessment_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'assessment id',
  `user_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'user id',
  `user_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'user type',

  `status_by_system` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'status_by_system',
  `status_by_user` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'status_by_user',

  `create_at` bigint(20) NOT NULL COMMENT 'create time (unix seconds)',
  `update_at` bigint(20) NOT NULL COMMENT 'update time (unix seconds)',
  `delete_at` bigint(20) NOT NULL COMMENT 'delete time (unix seconds)',
  PRIMARY KEY (`id`),
  KEY `assessment_id` (`assessment_id`),
  KEY `delete_at` (`delete_at`),

  UNIQUE `assessments_users_v2_unique` (`assessment_id`,`user_id`,`user_type`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='assessments_users_v2';

CREATE TABLE IF NOT EXISTS `assessments_contents_v2` (
     `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
     `assessment_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'assessment id',
     `content_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'content id',
     `content_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'content type',
     `status` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'status',
     `reviewer_comment` text COLLATE utf8mb4_unicode_ci default NULL COMMENT 'comment',

     `create_at` bigint(20) NOT NULL COMMENT 'create time (unix seconds)',
     `update_at` bigint(20) NOT NULL COMMENT 'update time (unix seconds)',
     `delete_at` bigint(20) NOT NULL COMMENT 'delete time (unix seconds)',
     PRIMARY KEY (`id`),
     KEY `assessment_id` (`assessment_id`),
     KEY `delete_at` (`delete_at`),
     UNIQUE `assessments_contents_v2_unique` (`assessment_id`,`content_id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='assessments_contents_v2';

CREATE TABLE IF NOT EXISTS `assessments_reviewer_feedback_v2` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `assessment_user_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'assessment user id',
  `status` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'process status',
  `reviewer_id` varchar(64) COLLATE utf8mb4_unicode_ci default NULL COMMENT 'reviewer id',
  `student_feedback_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL  COMMENT 'student latest feedback id',
  `assess_score` int(11) DEFAULT '0' COMMENT 'score',
  `complete_at` bigint(20) DEFAULT '0' COMMENT 'complete time (unix seconds)',
  `reviewer_comment` text COLLATE utf8mb4_unicode_ci default NULL COMMENT 'comment',

  `create_at` bigint(20) NOT NULL COMMENT 'create time (unix seconds)',
  `update_at` bigint(20) NOT NULL COMMENT 'update time (unix seconds)',
  `delete_at` bigint(20) NOT NULL COMMENT 'delete time (unix seconds)',
  PRIMARY KEY (`id`),
  KEY `assessment_user_id` (`assessment_user_id`),
  KEY `status` (`status`),
  KEY `delete_at` (`delete_at`),

  UNIQUE `assessments_users_results_v2_unique` (`assessment_user_id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='assessments_reviewer_feedback_v2';

CREATE TABLE IF NOT EXISTS `assessments_users_outcomes_v2` (
    `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
    `assessment_user_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'assessment user id',
    `assessment_content_id` varchar(64) COLLATE utf8mb4_unicode_ci default NULL COMMENT 'assessment content id',
    `outcome_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'content type',
    `status` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'status',

    `create_at` bigint(20) NOT NULL COMMENT 'create time (unix seconds)',
    `update_at` bigint(20) NOT NULL COMMENT 'update time (unix seconds)',
    `delete_at` bigint(20) NOT NULL COMMENT 'delete time (unix seconds)',
    PRIMARY KEY (`id`),

    UNIQUE `assessments_users_results_v2_unique` (`assessment_user_id`,`assessment_content_id`,`outcome_id`,`delete_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='assessments_users_outcomes_v2';