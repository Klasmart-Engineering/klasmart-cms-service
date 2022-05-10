update
    assessments_v2,
    (select distinct schedule_id,title from home_fun_studies_backup where delete_at=0) t1
set assessments_v2.title = t1.title
where
    assessments_v2.schedule_id = t1.schedule_id
  and assessments_v2.assessment_type='OfflineStudy';