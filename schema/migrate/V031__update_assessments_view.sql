-- assessments view
create or replace view assessments as
select id,schedule_id,title,class_end_at class_end_time,class_length,complete_at complete_time,
       if(status ='Complete','complete','in_progress') status,
       create_at,update_at,delete_at  from assessments_v2
where delete_at=0 and
    (
            (assessment_type in ('OfflineClass','OnlineClass') and status in ('Started','Draft','Complete'))
            or
            (assessment_type = 'OnlineStudy')
        );

-- assessments_attendances view
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
  and ((assessments_v2.assessment_type='OnlineClass' and assessments_users_v2.status_by_system='Participate')
    or assessments_v2.assessment_type in ('OfflineClass','OnlineStudy'));

-- assessments_outcomes view
create or replace view assessments_outcomes as
select
    assessments_users_outcomes_v2.id,
    assessments_users_v2.assessment_id,
    assessments_users_outcomes_v2.outcome_id,
    if(assessments_users_outcomes_v2.status='NotAchieved',1,0) none_achieved,
    if(assessments_users_outcomes_v2.status='NotCovered',1,0) skip,
    1 checked
from assessments_users_v2 inner join assessments_users_outcomes_v2
                                     on assessments_users_v2.id = assessments_users_outcomes_v2.assessment_user_id
where assessments_users_outcomes_v2.delete_at=0
group by assessments_users_v2.assessment_id,assessments_users_outcomes_v2.outcome_id;