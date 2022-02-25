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
where assessments_v2.delete_at = 0;