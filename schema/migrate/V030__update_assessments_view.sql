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