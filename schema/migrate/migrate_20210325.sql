ALTER TABLE assessments ADD (
    `lesson_plan_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'lesson plan id',
    `lesson_material_ids` JSON NOT NULL COMMENT 'lesson material ids'
)
ALTER TABLE assessments_outcomes ADD `checked` boolean NOT NULL DEFAULT true COMMENT 'checked'
ALTER TABLE assessments_attendances ADD (
    `origin` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'origin',
    `role` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'role'
)
