/* create table assessments_contents_outcomes_attendances */
CREATE TABLE IF NOT EXISTS `contents_outcomes_attendances` (
  `id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'id',
  `assessment_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'assessment id',
  `content_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'content id',
  `outcome_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'outcome id',
  `attendance_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'attendance id',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_assessment_id_content_id_outcome_id_attendance_id` (`assessment_id`,`content_id`,`outcome_id`, `attendance_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='assessment content outcome attendances (add: 2021-08-25)';

/* add none_achieved column for assessments_contents_outcomes table */
ALTER TABLE assessments_contents_outcomes ADD COLUMN none_achieved BOOLEAN NOT NULL DEFAULT false COMMENT 'none achieved (add: 2021-08-25)';
