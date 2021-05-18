ALTER TABLE assessments ADD `type` VARCHAR(1024) NOT NULL DEFAULT '' COMMENT 'assessment type';
UPDATE assessments SET type = 'class_and_live_outcome' where 'type' = '';
