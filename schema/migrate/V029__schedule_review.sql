CREATE TABLE IF NOT EXISTS `schedules_reviews` (
  `schedule_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'schedule_id',
  `student_id` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'student_id',
  `review_status` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'review_status',
  `live_lesson_plan` json COMMENT 'review lesson plan'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='schedules_reviews';

ALTER TABLE schedules ADD is_review tinyint(1) NOT NULL DEFAULT '0' COMMENT 'is_review';
ALTER TABLE schedules ADD review_status varchar(100) NOT NULL DEFAULT '' COMMENT 'review_status';
ALTER TABLE schedules ADD content_start_at bigint(20) NOT NULL DEFAULT '0' COMMENT 'content_start_at';
ALTER TABLE schedules ADD content_end_at bigint(20) NOT NULL DEFAULT '0' COMMENT 'content_end_at';

ALTER TABLE schedules_reviews ADD `type` varchar(100) NOT NULL DEFAULT '' COMMENT 'review_type';
ALTER TABLE schedules_reviews ADD UNIQUE INDEX uk_schedule_id_student_id ( schedule_id, student_id );