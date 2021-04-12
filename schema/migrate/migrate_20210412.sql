CREATE TABLE assessments_contents (
    `id` VARCHAR(64) NOT NULL COMMENT 'id',
    `assessment_id` VARCHAR(64) NOT NULL COMMENT 'assessment id',
    `content_id` VARCHAR(64) NOT NULL COMMENT 'content id',
    `content_name` VARCHAR(255) NOT NULL COMMENT 'content name',
    `content_type` int NOT NULL COMMENT 'content type',
    `content_comment` BOOLEAN NOT NULL COMMENT 'content comment',
    `checked` BOOLEAN NOT NULL DEFAULT true COMMENT 'checked',
    `outcome_ids` JSON NOT NULL COMMENT 'outcome ids',
    PRIMARY KEY (`id`),
    UNIQUE `uq_assessments_contents_assessment_id_content_id` (`assessment_id`, `content_id`),
    KEY `idx_assessments_contents_assessment_id` (`assessment_id`)
)  COMMENT 'assessment and outcome map' DEFAULT CHARSET=UTF8MB4 COLLATE = UTF8MB4_UNICODE_CI;

ALTER TABLE assessments_outcomes ADD `checked` boolean NOT NULL DEFAULT true COMMENT 'checked';

ALTER TABLE assessments_attendances ADD (
    `origin` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'origin',
    `role` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'role'
);
