alter table assessments rename to assessments_backup;
alter table assessments_attendances rename to assessments_attendances_backup;
alter table assessments_contents rename to assessments_contents_backup;
alter table assessments_outcomes rename to assessments_outcomes_backup;
alter table assessments_contents_outcomes rename to assessments_contents_outcomes_backup;
alter table home_fun_studies rename to home_fun_studies_backup;
alter table outcomes_attendances rename to outcomes_attendances_backup;
alter table contents_outcomes_attendances rename to contents_outcomes_attendances_backup;

create or replace view assessments as
select id,schedule_id,title,class_end_at class_end_time,class_length,complete_at complete_time,
       if(status ='Complete','complete','in_progress') status,
       create_at,update_at,delete_at  from assessments_v2
where delete_at=0 and
    (
            (assessment_type in ('OfflineClass','OnlineClass') and status in ('Started' 'Draft','Complete'))
            or
            (assessment_type = 'OnlineStudy')
        );

create or replace view assessments_attendances as
select id,assessment_id,
       user_id attendance_id,
       if(status_by_user='Participate',1,0) checked,
       'participants' origin,
       if(user_type='Teacher','teacher','student') role
from assessments_users_v2
where delete_at = 0;

create or replace view assessments_contents as
select
    id,
    assessment_id,
    content_id,
    '' content_name,
    if (content_type='LessonPlan',2,1) content_type,
    reviewer_comment content_comment,
    if(status='Covered',1,0) checked
from assessments_contents_v2
where delete_at = 0;

create or replace view assessments_outcomes as
select
    assessments_users_outcomes_v2.id,
    assessments_users_v2.assessment_id,
    assessments_users_outcomes_v2.outcome_id,
    if(assessments_users_outcomes_v2.status='NotAchieved',1,0) none_achieved,
    if(assessments_users_outcomes_v2.status='NotCovered',1,0) skip,
    if(assessments_users_outcomes_v2.status='Unknown',0,1) checked
from assessments_users_v2 inner join assessments_users_outcomes_v2
                                     on assessments_users_v2.id = assessments_users_outcomes_v2.assessment_user_id
where assessments_users_outcomes_v2.delete_at=0
group by assessments_users_v2.assessment_id,assessments_users_outcomes_v2.outcome_id;

create or replace view outcomes_attendances as
select
    assessments_users_outcomes_v2.id id,
    assessments_users_v2.assessment_id assessment_id,
    assessments_users_outcomes_v2.outcome_id outcome_id,
    assessments_users_v2.user_id attendance_id
from assessments_users_outcomes_v2 inner join assessments_users_v2
                                              on assessments_users_v2.id = assessments_users_outcomes_v2.assessment_user_id
where assessments_users_outcomes_v2.status = 'Achieved' and assessments_users_outcomes_v2.delete_at=0
group by assessments_users_v2.assessment_id,assessments_users_outcomes_v2.outcome_id,assessments_users_v2.user_id;

create or replace view assessments_contents_outcomes as
select
    assessments_users_outcomes_v2.id id,
    assessments_contents_v2.assessment_id assessment_id,
    assessments_users_outcomes_v2.outcome_id outcome_id,
    assessments_contents_v2.content_id content_id,
    if(assessments_users_outcomes_v2.status='NotAchieved',1,0) none_achieved
from assessments_users_outcomes_v2 inner join assessments_contents_v2
                                              on assessments_users_outcomes_v2.assessment_content_id = assessments_contents_v2.id
where assessments_users_outcomes_v2.delete_at=0
group by assessments_contents_v2.assessment_id,assessments_users_outcomes_v2.outcome_id,assessments_contents_v2.content_id;

create or replace view contents_outcomes_attendances as
select
    assessments_users_outcomes_v2.id id,
    assessments_contents_v2.assessment_id assessment_id,
    assessments_contents_v2.content_id content_id,
    assessments_users_outcomes_v2.outcome_id outcome_id,
    assessments_users_v2.user_id attendance_id
from assessments_users_outcomes_v2
         inner join assessments_users_v2
                    on assessments_users_outcomes_v2.assessment_user_id = assessments_users_v2.id
         inner join assessments_contents_v2
                    on assessments_users_outcomes_v2.assessment_content_id = assessments_contents_v2.id
where assessments_users_outcomes_v2.status='Achieved' and assessments_users_outcomes_v2.delete_at=0;

create or replace view home_fun_studies as
select
    assessments_reviewer_feedback_v2.id,
    assessments_v2.schedule_id,
    assessments_v2.title,
    '[]' teacher_ids,
    assessments_users_v2.user_id,
    if(assessments_reviewer_feedback_v2.status='Complete','complete','in_progress') status,
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