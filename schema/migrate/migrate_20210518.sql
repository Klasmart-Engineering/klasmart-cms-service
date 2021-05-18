ALTER TABLE assessments ADD `type` VARCHAR(1024) NOT NULL DEFAULT '' COMMENT 'assessment type';
ALTER TABLE assessments_contents ADD `content_source` TEXT NOT NULL DEFAULT '' COMMENT 'content source';
