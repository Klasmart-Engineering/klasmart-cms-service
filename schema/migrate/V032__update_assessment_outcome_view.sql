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