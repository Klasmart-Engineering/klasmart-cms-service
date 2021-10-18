-- return one row if schedule has non subject, and subject_id is ''
-- return one row if schedule has one subject, and subject_id has value
-- return n rows if schedule has n subject, and subject_id has value
create or replace view v_schedules_subjects as
(
select s.id                                           as schedule_id,
       sr1.relation_id                                as class_id,
       if(sr.relation_id is null, '', sr.relation_id) as subject_id
from schedules s
         left join schedules_relations sr on s.id = sr.schedule_id and sr.relation_type = 'Subject'
         left join schedules_relations sr1 on s.id = sr1.schedule_id and sr1.relation_type = 'class_roster_class'
where s.delete_at = 0
);

-- return m*n rows, if m students should achieve n outcomes
create or replace  view v_assessments_outcomes_students as
select
    t1.assessment_id,
    t1.outcome_id,
    t2.student_id,
    if(oa.id is null,0,1) as is_student_achieved
from
    (
        SELECT
            DISTINCT assessment_id,
                     outcome_id
        FROM
            assessments_outcomes
    ) t1
        INNER JOIN (
        SELECT
            assessment_id,
            attendance_id AS student_id
        FROM
            assessments_attendances
        WHERE
                checked = 1
          AND origin = 'class_roaster'
          AND role = 'student'
    ) t2 ON	t1.assessment_id = t2.assessment_id
        left join outcomes_attendances oa
            on t1.assessment_id = oa.assessment_id
            and t1.outcome_id = oa.outcome_id
            and t2.student_id = oa.attendance_id;