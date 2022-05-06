ALTER TABLE `schedules` ADD INDEX idx_program_id_org_id_delete_at ( `program_id`, `org_id`, `delete_at` );
DROP INDEX idx_org_id_delete_at_program_id ON schedules;