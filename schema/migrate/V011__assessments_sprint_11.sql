/* create table assessments_contents_outcomes_attendances */
CREATE TABLE IF NOT EXISTS `assessments_contents_outcomes_attendances` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'id',
  `assessment_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'assessment id',
  `content_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'content id',
  `outcome_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'outcome id',
  `attendance_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'attendance id',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_assessment_id_content_id_outcome_id_attendance_id` (`assessment_id`,`content_id`,`outcome_id`, `attendance_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='assessment content outcome attendances (add: 2021-08-20)';

/* add skip column for assessments_contents_outcomes table */
ALTER TABLE assessments_contents_outcomes ADD COLUMN skip BOOLEAN NOT NULL DEFAULT false COMMENT 'skip/not attempted (add: 2021-08-20)';

/* DELETE none_achieved column for assessments_outcomes table */
ALTER TABLE assessments_outcomes `none_achieved` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'none achieved (DELETED: 2021-08-20)',
