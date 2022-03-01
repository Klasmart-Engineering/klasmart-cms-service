update assessments_users_outcomes_v2,assessments_users_v2,outcomes_attendances_backup
set assessments_users_outcomes_v2.status = 'Achieved'
where assessments_users_outcomes_v2.assessment_user_id = assessments_users_v2.id
  and assessments_users_outcomes_v2.outcome_id = outcomes_attendances_backup.outcome_id
  and assessments_users_v2.user_id = outcomes_attendances_backup.attendance_id
  and assessments_users_v2.assessment_id = outcomes_attendances_backup.assessment_id;