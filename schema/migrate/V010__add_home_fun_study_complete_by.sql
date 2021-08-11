/* add column complete_by */
ALTER TABLE home_fun_studies ADD COLUMN complete_by VARCHAR(64) NULL DEFAULT '' comment 'complete user id (add: 2021-08-09)'

/* migrate data */
update home_fun_studies set complete_by = teacher_ids->>"$[0]" where complete_at = 0 and complete_by = ''
