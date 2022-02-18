-- home_fun_studies view
create or replace view home_fun_studies as
select
    assessments_reviewer_feedback_v2.id,
    assessments_v2.schedule_id,
    assessments_v2.title,
    '[]' teacher_ids,
    assessments_users_v2.user_id student_id,
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