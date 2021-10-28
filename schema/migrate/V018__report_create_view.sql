-- return one row if schedule has non subject, and subject_id is ''
-- return one row if schedule has one subject, and subject_id has value
-- return n rows if schedule has n subject, and subject_id has value
create or replace view v_schedules_subjects as
(
select sr.schedule_id,
       sr.relation_id                                as class_id,
       if(sr1.relation_id is null, '', sr1.relation_id) as subject_id
from   schedules_relations sr
         left join schedules_relations sr1 on sr.schedule_id = sr1.schedule_id and sr1.relation_type = 'Subject'
where   sr.relation_type = 'class_roster_class'
);
