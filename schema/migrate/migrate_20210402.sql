/* assessments_contents */
CREATE TABLE assessments_contents (
    `id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'id',
    `assessment_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'assessment id',
    `content_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'content id',
    `content_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'content name',
    `content_type` INT NOT NULL DEFAULT 0 COMMENT 'content type',
    `content_comment` TEXT NOT NULL COMMENT 'content comment',
    `checked` BOOLEAN NOT NULL DEFAULT true COMMENT 'checked',
    PRIMARY KEY (`id`),
    UNIQUE `uq_assessments_contents_assessment_id_content_id` (`assessment_id`, `content_id`),
    KEY `idx_assessments_contents_assessment_id` (`assessment_id`)
)  COMMENT 'assessment and outcome map' DEFAULT CHARSET=UTF8MB4 COLLATE = UTF8MB4_UNICODE_CI;

/* assessments_contents_outcomes */
CREATE TABLE assessments_contents_outcomes (
    `id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'id',
    `assessment_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'assessment id',
    `content_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'content id',
    `outcome_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'outcome id',
    PRIMARY KEY (`id`),
    UNIQUE `uq_assessments_contents_outcomes` (`assessment_id`, `content_id`, `outcome_id`)
) COMMENT 'assessment content and outcome map' DEFAULT CHARSET=UTF8MB4 COLLATE = UTF8MB4_UNICODE_CI;

/* assessments_outcomes */
ALTER TABLE assessments_outcomes ADD `checked` boolean NOT NULL DEFAULT true COMMENT 'checked';

/* assessments_attendances */
ALTER TABLE assessments_attendances ADD (
    `origin` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'origin',
    `role` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'role'
);
CREATE UNIQUE INDEX uq_assessments_attendances_assessment_id_attendance_id_role on assessments_attendances(assessment_id, attendance_id, role);

/* assessments */
ALTER TABLE assessments MODIFY `program_id` VARCHAR(64) NULL DEFAULT '' COMMENT 'DEPRECATED: program id';
ALTER TABLE assessments MODIFY `subject_id` VARCHAR(64) NULL DEFAULT '' COMMENT 'DEPRECATED: subject id';
ALTER TABLE assessments MODIFY `teacher_ids` JSON NULL COMMENT 'DEPRECATED: teacher ids';
