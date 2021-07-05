-- cms_contents
ALTER TABLE cms_contents MODIFY 
 `program` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'program';
 
ALTER TABLE cms_contents MODIFY 
 `subject` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'subject';
 
ALTER TABLE cms_contents MODIFY 
 `developmental` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'developmental';
 
 ALTER TABLE cms_contents MODIFY 
 `skills` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'skills';
 
 ALTER TABLE cms_contents MODIFY 
 `age` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'age';
 
 ALTER TABLE cms_contents MODIFY 
 `grade` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'grade';
 
 
-- learning_outcomes
 ALTER TABLE learning_outcomes MODIFY 
 `program` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'program';
 
ALTER TABLE learning_outcomes MODIFY 
 `subject` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'subject';
 
ALTER TABLE learning_outcomes MODIFY 
 `developmental` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'developmental';
 
 ALTER TABLE learning_outcomes MODIFY 
 `skills` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'skills';
 
 ALTER TABLE learning_outcomes MODIFY 
 `age` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'age';
 
 ALTER TABLE learning_outcomes MODIFY 
 `grade` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'grade';


 ALTER TABLE assessments_contents MODIFY
 `content_comment` text COLLATE utf8mb4_unicode_ci NULL COMMENT 'content comment';

ALTER TABLE learning_outcomes MODIFY `ancestor_id` VARCHAR ( 50 ) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'ancestor_id';
ALTER TABLE learning_outcomes MODIFY `shortcode` CHAR ( 20 ) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'shortcode';
ALTER TABLE learning_outcomes MODIFY `name` VARCHAR ( 255 ) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'outcome_name';
ALTER TABLE learning_outcomes MODIFY `estimated_time` INT ( 11 ) NOT NULL DEFAULT 0 COMMENT 'estimated_time';
ALTER TABLE learning_outcomes MODIFY `author_id` VARCHAR ( 50 ) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'author_id';
ALTER TABLE learning_outcomes MODIFY `author_name` VARCHAR ( 128 ) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'author_name';
ALTER TABLE learning_outcomes MODIFY `organization_id` VARCHAR ( 50 ) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'organization_id';
ALTER TABLE learning_outcomes MODIFY `publish_status` VARCHAR ( 16 ) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'publish_status';
ALTER TABLE learning_outcomes MODIFY `create_at` BIGINT ( 20 ) NOT NULL DEFAULT 0 COMMENT 'created_at';
ALTER TABLE learning_outcomes MODIFY `update_at` BIGINT ( 20 ) NOT NULL DEFAULT 0 COMMENT 'updated_at';

-- schedules
ALTER TABLE schedules MODIFY `title` VARCHAR ( 100 ) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'title';
ALTER TABLE schedules MODIFY `class_id` VARCHAR ( 100 ) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'class_id';
ALTER TABLE schedules MODIFY `org_id` VARCHAR ( 100 ) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'org_id';
ALTER TABLE schedules MODIFY `class_type` VARCHAR ( 100 ) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'class_type';
ALTER TABLE schedules MODIFY `start_at` BIGINT ( 20 ) NOT NULL DEFAULT 0 COMMENT 'start_at';
ALTER TABLE schedules MODIFY `end_at` BIGINT ( 20 ) NOT NULL DEFAULT 0 COMMENT 'end_at';