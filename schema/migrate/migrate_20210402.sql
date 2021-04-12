CREATE TABLE assessments_contents (
    `id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'id',
    `assessment_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'assessment id',
    `content_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'content id',
    `content_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'content name',
    `content_type` int NOT NULL DEFAULT '' COMMENT 'content type',
    `content_comment` BOOLEAN NOT NULL DEFAULT '' COMMENT 'content comment',
    `checked` BOOLEAN NOT NULL DEFAULT true COMMENT 'checked',
    `outcome_ids` JSON NOT NULL DEFAULT '' COMMENT 'outcome ids',
    PRIMARY KEY (`id`),
    UNIQUE `uq_assessments_contents_assessment_id_content_id` (`assessment_id`, `content_id`),
    KEY `idx_assessments_contents_assessment_id` (`assessment_id`)
)  COMMENT 'assessment and outcome map' DEFAULT CHARSET=UTF8MB4 COLLATE = UTF8MB4_UNICODE_CI;

ALTER TABLE assessments_outcomes ADD `checked` boolean NOT NULL DEFAULT true COMMENT 'checked';

ALTER TABLE assessments_attendances ADD (
    `origin` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'origin',
    `role` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'role'
);
CREATE UNIQUE INDEX uq_assessments_attendances_assessment_id_attendance_id on assessments_attendances(assessment_id, attendance_id);