/* remove duplicate outcome attendance records */
create temporary table `temp_deleting_outcomes_attendances`(
  select
    id
  from
    `outcomes_attendances`
  where
    (`assessment_id`, `outcome_id`, `attendance_id`) in (
      select
        `assessment_id`,
        `outcome_id`,
        `attendance_id`
      from
        outcomes_attendances
      group by
        `assessment_id`,
        `outcome_id`,
        `attendance_id`
      having
        count(*) > 1
    )
    and id not in (
      select
        min(id)
      from
        outcomes_attendances
      group by
        `assessment_id`,
        `outcome_id`,
        `attendance_id`
      having
        count(*) > 1
    )
);

delete from
  `outcomes_attendances`
where
  id in (
    select
      id
    from
      temp_deleting_outcomes_attendances
  );

drop table `temp_deleting_outcomes_attendances`;

/* migrate outcomes_attendances data to contents_outcomes_attendances */
insert into
  contents_outcomes_attendances (
    id,
    assessment_id,
    outcome_id,
    content_id,
    attendance_id
  )
select
  CONCAT("0000", UUID_SHORT()) as id,
  t.assessment_id,
  t.outcome_id,
  t.content_id,
  t.attendance_id
from
  (
    select
      distinct outcomes_attendances.assessment_id,
      outcomes_attendances.outcome_id,
      assessments_contents_outcomes.content_id,
      outcomes_attendances.attendance_id
    from
      outcomes_attendances
      inner join assessments_contents_outcomes on outcomes_attendances.assessment_id = assessments_contents_outcomes.assessment_id
      and outcomes_attendances.outcome_id = assessments_contents_outcomes.outcome_id
      inner join assessments_outcomes on outcomes_attendances.assessment_id = assessments_outcomes.assessment_id
      and outcomes_attendances.outcome_id = assessments_outcomes.outcome_id
    where
      (
        outcomes_attendances.assessment_id,
        outcomes_attendances.outcome_id
      ) not in (
        select
          assessment_id,
          outcome_id
        from
          contents_outcomes_attendances
      )
      and assessments_outcomes.checked = 1
  ) t;