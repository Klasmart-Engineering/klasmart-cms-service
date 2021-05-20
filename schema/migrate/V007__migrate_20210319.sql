CREATE UNIQUE INDEX uq_home_fun_studies_schedule_id_and_student_id ON home_fun_studies(schedule_id, student_id);
ALTER TABLE home_fun_studies RENAME INDEX `home_fun_study_id` TO `idx_home_fun_studies_schedule_id`;
ALTER TABLE home_fun_studies RENAME INDEX `home_fun_study_status` TO `idx_home_fun_studies_status`;
ALTER TABLE home_fun_studies RENAME INDEX `home_fun_study_latest_feedback_at` TO `idx_home_fun_studies_latest_feedback_at`;
ALTER TABLE home_fun_studies RENAME INDEX `home_fun_study_complete_at` TO `idx_home_fun_studies_complete_at`;
