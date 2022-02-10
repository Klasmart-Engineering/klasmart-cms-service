-- assessment data 
INSERT IGNORE INTO
    assessments_v2(id,org_id,schedule_id,assessment_type,title,status,complete_at,class_length,class_end_at,create_at,update_at,delete_at,migrate_flag)
SELECT
    IFNULL(assessments.id,schedules.id) id,
    schedules.org_id,
    schedules.id schedule_id,
    IF(schedules.class_type='Homework' AND schedules.is_home_fun=0,'OnlineStudy',
       IF(schedules.class_type='Homework' AND schedules.is_home_fun=1,'OfflineStudy',schedules.class_type)) assessment_type,
    IFNULL(assessments.title,schedules.title) title,
    CASE
        WHEN schedules.class_type='Homework' AND schedules.is_home_fun=1 THEN 'NA'
        WHEN (assessments.id IS NULL AND schedules.live_lesson_plan IS NULL) THEN 'NotStarted'
        ELSE assessments.status END status,
    IFNULL(assessments.complete_time,0) complete_time,
    IFNULL(assessments.class_length,0) class_length,
    IFNULL(assessments.class_end_time,0) class_end_time,
    IF(assessments.create_at=0 OR assessments.create_at IS NULL,schedules.created_at,assessments.create_at) create_at,
    IFNULL(assessments.update_at,0) update_at,
    0 delete_at,
    1 migrate_flag
FROM schedules
         LEFT JOIN assessments ON schedules.id = assessments.schedule_id AND assessments.delete_at=0
WHERE schedules.delete_at=0 AND schedules.class_type!='Task';

update assessments_v2 set status ='Complete' WHERE status = 'complete';
update assessments_v2 set status ='Started' WHERE status = 'in_progress';

-- assessment user
INSERT IGNORE INTO
    assessments_users_v2(id,assessment_id,user_id,user_type,status_by_system,status_by_user,create_at,update_at,delete_at)
SELECT
    schedules_relations.id                                                                                                                      id,
    assessments_v2.id                                                                                                                           assessment_id,
    schedules_relations.relation_id                                                                                                             user_id,
    IF(schedules_relations.relation_type='class_roster_teacher' OR schedules_relations.relation_type='participant_teacher','Teacher','Student') user_type,
    IF(assessments_attendances.checked=1,'Participate','NotParticipate')                                                                        status_by_system,
    IF(assessments_v2.assessment_type = 'OnlineClass',
       IF(assessments_attendances.checked = 1, 'Participate', 'NotParticipate'),
       IF(assessments_attendances.checked IS NULL OR assessments_attendances.checked = 1, 'Participate',
          'NotParticipate'))                                                                                                                    status_by_user,
    assessments_v2.create_at                                                                                                                    create_at,
    0                                                                                                                                           update_at,
    0                                                                                                                                           delete_at
FROM schedules_relations
         INNER JOIN assessments_v2 ON schedules_relations.schedule_id = assessments_v2.schedule_id
         LEFT JOIN  assessments_attendances
                    ON assessments_v2.id=assessments_attendances.assessment_id AND schedules_relations.relation_id = assessments_attendances.attendance_id
WHERE schedules_relations.relation_type in ('class_roster_teacher','participant_teacher','class_roster_student','participant_student');

-- assessment content
INSERT IGNORE INTO
    assessments_contents_v2(id,assessment_id,content_id,content_type,status,reviewer_comment,create_at,update_at,delete_at)
SELECT
    assessments_contents.id,
    assessments_contents.assessment_id,
    assessments_contents.content_id,
    IF(assessments_contents.content_type=2,'LessonPlan','LessonMaterial') content_type,
    IF(assessments_contents.checked=1,'Covered','NotCovered') status,
    assessments_contents.content_comment reviewer_comment,
    assessments_v2.create_at create_at,
    0 update_at,
    0 delete_at
FROM assessments_v2
         INNER JOIN assessments_contents
                    ON assessments_v2.id = assessments_contents.assessment_id;

-- assessment user result/ home_fun_studies
INSERT IGNORE INTO
    assessments_reviewer_feedback_v2(id,assessment_user_id,status,reviewer_id,student_feedback_id,assess_score,complete_at,reviewer_comment,create_at,update_at,delete_at)
SELECT
    home_fun_studies.id,
    assessments_users_v2.id assessment_user_id,
    home_fun_studies.status,
    home_fun_studies.complete_by,
    home_fun_studies.assess_feedback_id,
    home_fun_studies.assess_score,
    home_fun_studies.complete_at,
    home_fun_studies.assess_comment,
    home_fun_studies.create_at,
    home_fun_studies.update_at,
    0 delete_at
FROM assessments_v2
         INNER JOIN assessments_users_v2
                    ON assessments_v2.id = assessments_users_v2.assessment_id
         INNER JOIN home_fun_studies
                    ON assessments_v2.schedule_id = home_fun_studies.schedule_id
                        AND assessments_users_v2.user_id = home_fun_studies.student_id
WHERE home_fun_studies.delete_at = 0;

update assessments_reviewer_feedback_v2 set status='Complete' where status='complete';
update assessments_reviewer_feedback_v2 set status='Started' where status='in_progress';


-- assessment user outcome
DROP TABLE
    IF
        EXISTS tmp_assessments_users_outcomes_v2;
CREATE TABLE tmp_assessments_users_outcomes_v2 LIKE assessments_users_outcomes_v2;

ALTER TABLE tmp_assessments_users_outcomes_v2
    ADD COLUMN assessment_id varchar(255) DEFAULT NULL,
    ADD COLUMN user_id varchar(255) DEFAULT NULL,
    ADD COLUMN content_id varchar(255) DEFAULT NULL,
    ADD COLUMN schedule_id varchar(255) DEFAULT NULL;

DROP TABLE
    IF
        EXISTS tmp_cms_contents_outcomes;
CREATE TABLE tmp_cms_contents_outcomes(
                                          `content_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
                                          `outcome_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO  tmp_cms_contents_outcomes(content_id,outcome_id)
SELECT cms_contents.id,
       substring_index(substring_index(cms_contents.outcomes, ',', b.id + 1 ), ',',- 1) outcome_id
FROM cms_contents
         JOIN
     (select 0 as id union all select 1 union all select 2 union all select 3 union all select 4 union all select 5
      union all
      select 6 union all select 7 union all select 8 union all select 9 union all select 10
      union all
      select 11 union all select 12 union all select 13 union all select 14 union all select 15
      union all
      select 16 union all select 17 union all select 18 union all select 19 union all select 20
      union all
      select 21 union all select 22 union all select 23 union all select 24 union all select 25) b
     ON b.id < (length( cms_contents.outcomes )-length( REPLACE (cms_contents.outcomes, ',', '')) + 1)
WHERE cms_contents.outcomes !='' AND cms_contents.outcomes IS NOT NULL;

INSERT IGNORE INTO
    tmp_assessments_users_outcomes_v2
( id,
  assessment_user_id,
  assessment_content_id,
  outcome_id,
  status,create_at,
  update_at,
  delete_at,
  assessment_id,
  user_id,
  content_id,
  schedule_id
)
SELECT
    replace(uuid(),"-","") as uuid,
    assessments_users_v2.id assessment_user_id,
    assessments_contents_v2.id assessment_content_id,
    tmp_cms_contents_outcomes.outcome_id,
    'Unknown' status,
    assessments_v2.create_at,
    0 update_at,
    0 delete_at,
    assessments_v2.id,
    assessments_users_v2.user_id,
    assessments_contents_v2.content_id,
    assessments_v2.schedule_id
FROM assessments_v2
         INNER JOIN assessments_users_v2
                    ON assessments_v2.id = assessments_users_v2.assessment_id
         LEFT JOIN  assessments_contents_v2
                    ON  assessments_v2.id = assessments_contents_v2.assessment_id
         LEFT JOIN
     tmp_cms_contents_outcomes ON  assessments_contents_v2.content_id = tmp_cms_contents_outcomes.content_id
WHERE
    tmp_cms_contents_outcomes.outcome_id IS NOT NULL
  AND assessments_v2.status in ('Started','Draft','Complete')
  AND assessments_users_v2.user_type='Student';

INSERT IGNORE INTO assessments_users_outcomes_v2
(id,assessment_user_id,assessment_content_id,outcome_id,status,create_at,update_at,delete_at)
SELECT
    tmp_assessments_users_outcomes_v2.id,
    tmp_assessments_users_outcomes_v2.assessment_user_id,
    tmp_assessments_users_outcomes_v2.assessment_content_id,
    tmp_assessments_users_outcomes_v2.outcome_id,
    IF(assessments_outcomes.skip=1,"NotCovered",
       IF(assessments_outcomes.none_achieved=1,"NotAchieved",
          IF(contents_outcomes_attendances.id IS NOT NULL,'Achieved','Unknown'))) status,
    tmp_assessments_users_outcomes_v2.create_at,
    0 update_at,
    0 delete_at
FROM tmp_assessments_users_outcomes_v2
         LEFT JOIN
     assessments_outcomes
     ON assessments_outcomes.assessment_id = tmp_assessments_users_outcomes_v2.assessment_id
         AND assessments_outcomes.outcome_id = tmp_assessments_users_outcomes_v2.outcome_id
         LEFT JOIN contents_outcomes_attendances
                   ON contents_outcomes_attendances.assessment_id = tmp_assessments_users_outcomes_v2.assessment_id
                       AND contents_outcomes_attendances.outcome_id = tmp_assessments_users_outcomes_v2.outcome_id
                       AND contents_outcomes_attendances.attendance_id = tmp_assessments_users_outcomes_v2.user_id
                       AND contents_outcomes_attendances.content_id = tmp_assessments_users_outcomes_v2.content_id
WHERE
        assessments_outcomes.outcome_id !=''  AND assessments_outcomes.outcome_id IS NOT NULL;

INSERT IGNORE INTO  tmp_assessments_users_outcomes_v2
(id,assessment_user_id,assessment_content_id,outcome_id,status,create_at,update_at,delete_at,assessment_id,user_id,schedule_id)
SELECT replace(uuid(),"-","") as uuid,
       assessments_users_v2.id assessment_user_id,
       null,
       schedules_relations.relation_id,
       'Unknown' status,
       assessments_v2.create_at,
       0 update_at,
       0 delete_at,
       assessments_v2.id,
       assessments_users_v2.user_id,
       assessments_v2.schedule_id
FROM assessments_v2
         LEFT JOIN
     assessments_users_v2
     ON assessments_v2.id = assessments_users_v2.assessment_id
         LEFT JOIN schedules_relations ON assessments_v2.schedule_id = schedules_relations.schedule_id
WHERE schedules_relations.relation_type='learning_outcome';

INSERT IGNORE INTO assessments_users_outcomes_v2
(id,assessment_user_id,assessment_content_id,outcome_id,status,create_at,update_at,delete_at)
SELECT
    tmp_assessments_users_outcomes_v2.id,
    tmp_assessments_users_outcomes_v2.assessment_user_id,
    tmp_assessments_users_outcomes_v2.assessment_content_id,
    tmp_assessments_users_outcomes_v2.outcome_id,
    IF(assessments_outcomes.skip=1,"NotCovered",
       IF(assessments_outcomes.none_achieved=1,"NotAchieved",
          IF(outcomes_attendances.id IS NOT NULL,'Achieved','Unknown'))) status,
    tmp_assessments_users_outcomes_v2.create_at,
    0 update_at,
    0 delete_at
FROM
    tmp_assessments_users_outcomes_v2
        INNER JOIN home_fun_studies
                   ON tmp_assessments_users_outcomes_v2.schedule_id = home_fun_studies.schedule_id AND tmp_assessments_users_outcomes_v2.user_id=home_fun_studies.student_id
        LEFT JOIN
    assessments_outcomes
    ON assessments_outcomes.assessment_id = home_fun_studies.id
        AND assessments_outcomes.outcome_id = tmp_assessments_users_outcomes_v2.outcome_id
        LEFT JOIN outcomes_attendances
                  ON outcomes_attendances.assessment_id = home_fun_studies.id
                      AND outcomes_attendances.outcome_id = tmp_assessments_users_outcomes_v2.outcome_id
                      AND outcomes_attendances.attendance_id = tmp_assessments_users_outcomes_v2.user_id;


DROP TABLE
    IF EXISTS tmp_assessments_users_outcomes_v2;
DROP TABLE
    IF EXISTS tmp_cms_contents_outcomes;
