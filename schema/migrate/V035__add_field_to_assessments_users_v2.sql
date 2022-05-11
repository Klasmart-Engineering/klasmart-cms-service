alter table assessments_users_v2 add in_progress_at bigint(20) NOT NULL DEFAULT 0 COMMENT 'in progress time (unix seconds)';
alter table assessments_users_v2 add done_at bigint(20) NOT NULL DEFAULT 0 COMMENT 'done time (unix seconds)';
alter table assessments_users_v2 add resubmitted_at bigint(20) NOT NULL DEFAULT 0 COMMENT 'resubmitted time (unix seconds)';
alter table assessments_users_v2 add completed_at bigint(20) NOT NULL DEFAULT 0 COMMENT 'completed time (unix seconds)';

-- update offline study assessments_users
update assessments_users_v2,assessments_reviewer_feedback_v2
set assessments_users_v2.done_at = assessments_reviewer_feedback_v2.create_at
where
    assessments_users_v2.id = assessments_reviewer_feedback_v2.assessment_user_id
  and
    assessments_users_v2.status_by_system = 'Done'
  and
    assessments_users_v2.id <>'';

-- update Resubmitted status assessments_users
update assessments_users_v2
set resubmitted_at = update_at
where assessments_users_v2.status_by_system = 'Resubmitted'
  and
        assessments_users_v2.id <>'';

-- update Completed status assessments_users
update assessments_users_v2,assessments_v2
set assessments_users_v2.completed_at =  if(assessments_v2.complete_at<>0,assessments_v2.complete_at,assessments_v2.create_at)
where
    assessments_users_v2.assessment_id = assessments_v2.id
  and
    assessments_users_v2.status_by_system = 'Completed'
  and
    assessments_users_v2.id <>'';

-- update OnlineClass,OfflineClass,OnlineStudy,ReviewStudy assessments_users
update assessments_users_v2,assessments_v2
set assessments_users_v2.done_at = assessments_v2.create_at
where
    assessments_v2.assessment_type in ('OnlineClass','OfflineClass','OnlineStudy','ReviewStudy')
  and
    assessments_users_v2.assessment_id = assessments_v2.id
  and
    assessments_users_v2.status_by_system = 'Done'
  and
    assessments_users_v2.id <>'';






