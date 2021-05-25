/* add assessments filed: type */
ALTER TABLE assessments ADD `type` VARCHAR(1024) NOT NULL DEFAULT '' COMMENT 'assessment type';
/* update assessment field: type = 'class_and_live_outcome' */
UPDATE assessments SET `type` = 'class_and_live_outcome' where `type` = '';
