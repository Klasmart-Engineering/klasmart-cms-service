-- update OfflineStudy status_by_system data in assessments_users_v2 table
update assessments_users_v2,
    (select
    assessments_users_v2.id,
    if(assessments_reviewer_feedback_v2.status is null,'NotStarted',
    if(assessments_v2.status = 'Complete','Completed',
    if(assessments_reviewer_feedback_v2.status in ('Started','Draft','Complete'),'Done',assessments_reviewer_feedback_v2.status))) status
    from
    assessments_v2
    inner join assessments_users_v2
    on assessments_v2.id = assessments_users_v2.assessment_id
    left join  assessments_reviewer_feedback_v2
    on assessments_users_v2.id = assessments_reviewer_feedback_v2.assessment_user_id
    where assessments_v2.assessment_type = 'OfflineStudy'
    and assessments_v2.delete_at = 0
    ) t1
set assessments_users_v2.status_by_system = t1.status
where
    assessments_users_v2.id = t1.id
  and assessments_users_v2.id <>"";

-- update OnlineClass,OfflineClass,OnlineStudy,ReviewStudy status_by_system data in assessments_users_v2 table
update assessments_users_v2,
    (select
    assessments_users_v2.id id,
    if(assessments_users_v2.status_by_system='Participate' and assessments_v2.status='Complete'
    ,'Completed',if(assessments_users_v2.status_by_system='Participate','Done',
    if(assessments_users_v2.status_by_system='NotParticipate','NotStarted',assessments_users_v2.status_by_system))) status
    from
    assessments_v2
    inner join assessments_users_v2
    on assessments_v2.id = assessments_users_v2.assessment_id
    where assessments_v2.assessment_type in ('OnlineClass','OfflineClass','OnlineStudy','ReviewStudy')
    and assessments_v2.delete_at = 0) t1
set assessments_users_v2.status_by_system = t1.status
where assessments_users_v2.id = t1.id
  and assessments_users_v2.id <>"";

-- update home_fun_studies view
create or replace view home_fun_studies as
select
    assessments_reviewer_feedback_v2.id,
    assessments_v2.schedule_id,
    assessments_v2.title,
    '[]' teacher_ids,
    assessments_users_v2.user_id student_id,
    if(assessments_v2.status='Complete','complete','in_progress') status,
    0 due_at,
    assessments_reviewer_feedback_v2.complete_at complete_at,
    assessments_reviewer_feedback_v2.student_feedback_id latest_feedback_id,
    assessments_reviewer_feedback_v2.student_feedback_id assess_feedback_id,
    0 latest_feedback_at,
    assessments_reviewer_feedback_v2.assess_score assess_score,
    assessments_reviewer_feedback_v2.reviewer_comment assess_comment,
    assessments_reviewer_feedback_v2.reviewer_id complete_by,
    assessments_reviewer_feedback_v2.create_at,
    assessments_reviewer_feedback_v2.update_at,
    assessments_reviewer_feedback_v2.delete_at
from assessments_reviewer_feedback_v2
         inner join assessments_users_v2
                    on assessments_reviewer_feedback_v2.assessment_user_id = assessments_users_v2.id
         inner join assessments_v2
                    on assessments_users_v2.assessment_id = assessments_v2.id
where assessments_reviewer_feedback_v2.delete_at=0;

-- update assessments_attendances view
create or replace view assessments_attendances as
select
    assessments_users_v2.id,
    assessments_users_v2.assessment_id,
    assessments_users_v2.user_id attendance_id,
    if(assessments_users_v2.status_by_user='Participate',1,0) checked,
    if(schedules_relations.relation_type in ('class_roster_teacher','class_roster_student'),'class_roaster','participants') origin,
    if(assessments_users_v2.user_type='Teacher','teacher','student') role
from assessments_users_v2
         inner join assessments_v2
                    on assessments_users_v2.assessment_id = assessments_v2.id
         inner join schedules_relations
                    on assessments_v2.schedule_id = schedules_relations.schedule_id and assessments_users_v2.user_id = schedules_relations.relation_id
where assessments_v2.delete_at = 0
  and ((assessments_v2.assessment_type='OnlineClass' and assessments_users_v2.status_by_system in ('InProgress','Done','Resubmitted','Completed'))
    or (assessments_v2.assessment_type = 'OfflineClass' and assessments_v2.status in ('Started','Draft','Complete'))
    or (assessments_v2.assessment_type = 'OnlineStudy'));

-- take the complete_at of the last student as the complete_at of the entire assessment
update
    assessments_v2,
    (select
    assessments_users_v2.assessment_id,
    max(assessments_reviewer_feedback_v2.complete_at) complete_at
    from
    assessments_users_v2
    inner join
    assessments_reviewer_feedback_v2
    on
    assessments_users_v2.id = assessments_reviewer_feedback_v2.assessment_user_id
    where assessments_reviewer_feedback_v2.status='Complete'
    group by assessments_users_v2.assessment_id) t1
set assessments_v2.complete_at = t1.complete_at
where assessments_v2.id = t1.assessment_id
  and assessments_v2.complete_at=0
  and assessments_v2.status = 'Complete';

-- create user_id index in assessments_users_v2 table
CREATE INDEX user_id ON assessments_users_v2 (user_id);