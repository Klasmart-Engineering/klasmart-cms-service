/* remove duplicate outcome attendance records */
DELETE FROM `outcomes_attendances` WHERE id IN (
    SELECT id
    FROM `outcomes_attendances`
    WHERE
    (`assessment_id`,`outcome_id`,`attendance_id`) IN (
        SELECT `assessment_id`,`outcome_id`,`attendance_id` FROM outcomes_attendances
        GROUP BY `assessment_id`,`outcome_id`,`attendance_id`
        HAVING COUNT(*) > 1
    )
    AND id NOT IN
    (
        SELECT MIN(id) FROM outcomes_attendances GROUP BY `assessment_id`,`outcome_id`,`attendance_id` HAVING COUNT(*) > 1
    )
)

