/* add column complete_by */
ALTER TABLE home_fun_studies ADD COLUMN complete_by VARCHAR(64) NOT NULL DEFAULT '' comment 'complete user id (add: 2021-08-09)';

/* migrate data */
-- update home_fun_studies set complete_by = teacher_ids->>"$[0]" where status = 'complete' and complete_by = '';
update home_fun_studies set complete_by = JSON_UNQUOTE(JSON_EXTRACT(teacher_ids,"$[0]")) where status = 'complete' and complete_by = '';
