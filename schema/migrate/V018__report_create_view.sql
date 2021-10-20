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

-- return the time student first achieve outcome
create or replace view  v_students_outcomes_first_achieve as
select
    vaos.student_id,vaos.outcome_id,
    min(a.complete_time) as first_achieve_time
from
    v_assessments_outcomes_students vaos
        inner join assessments a on
            a.id = vaos.assessment_id
where a.status ='complete' and vaos.is_student_achieved=1
group by vaos.student_id,vaos.outcome_id;