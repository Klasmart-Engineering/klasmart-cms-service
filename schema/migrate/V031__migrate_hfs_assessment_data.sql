-- add review_attachment_id field
alter table feedbacks_assignments add review_attachment_id varchar(500) DEFAULT NULL COMMENT 'attachment id for teacher evaluation';

-- migrate hsf assessment data
update assessments_v2,
(select
assessments_v2.id id,
if(sum(if(assessments_reviewer_feedback_v2.status='Complete',0,1))=0,'Complete',
if(sum(if(assessments_reviewer_feedback_v2.status is null,0,1))=0,'NotStarted','Started')) status
from
assessments_v2
left join
assessments_users_v2
on assessments_v2.id = assessments_users_v2.assessment_id
left join
assessments_reviewer_feedback_v2
on assessments_users_v2.id = assessments_reviewer_feedback_v2.assessment_user_id
where
assessments_v2.assessment_type = "OfflineStudy"
and assessments_v2.delete_at=0
and assessments_users_v2.user_type='Student'
group by assessments_v2.id) t2
set assessments_v2.status = t2.status
where assessments_v2.id=t2.id;
