ALTER TABLE schedules_reviews ADD `type` varchar(100) NOT NULL DEFAULT '' COMMENT 'review_type';

ALTER TABLE schedules_reviews ADD UNIQUE INDEX uk_schedule_id_student_id ( schedule_id, student_id );