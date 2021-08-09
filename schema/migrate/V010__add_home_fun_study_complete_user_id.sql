/* add column complete_user_id */
ALTER TABLE home_fun_studies ADD COLUMN complete_user_id VARCHAR(64) NULL comment 'complete user id (add: 2021-08-09)'

/* migrate data */
update home_fun_studies set complete_user_id = teacher_ids->>"$[0]" where complete_at = 0 and complete_user_id = ''
