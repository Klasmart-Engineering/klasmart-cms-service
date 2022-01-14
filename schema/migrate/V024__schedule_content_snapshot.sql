-- remove deduplicates
DROP TABLE
IF
	EXISTS tmp_assessments;
CREATE TABLE tmp_assessments LIKE assessments;
INSERT INTO tmp_assessments SELECT
* 
FROM
	assessments;
TRUNCATE TABLE assessments;
ALTER TABLE assessments ADD UNIQUE INDEX uk_schedule_id_delete_at ( schedule_id, delete_at );
INSERT IGNORE INTO assessments SELECT
* 
FROM
	tmp_assessments 
WHERE
	STATUS = 'complete' 
	AND delete_at = 0;
INSERT IGNORE INTO assessments SELECT
* 
FROM
	tmp_assessments;

-- add tmp columns
ALTER TABLE schedules ADD live_lesson_plan json;
ALTER TABLE schedules ADD tmp_content_id VARCHAR ( 64 );
ALTER TABLE schedules ADD tmp_content_name VARCHAR ( 255 );

-- insert live_lesson_plan
INSERT INTO schedules ( id, tmp_content_id, tmp_content_name ) SELECT
*
FROM
	(
	SELECT
		assessments.schedule_id AS id,
		assessments_contents.content_id AS content_id,
		cms_contents.content_name AS content_name 
	FROM
		assessments_contents
		INNER JOIN assessments ON assessments_contents.assessment_id = assessments.id
		LEFT JOIN cms_contents ON assessments_contents.content_id = cms_contents.id 
	WHERE
		assessments_contents.content_type = 2
	) AS t 
	ON DUPLICATE KEY UPDATE live_lesson_plan = JSON_SET(
		"{}",
		'$.lesson_plan_id',
		t.content_id,
		'$.lesson_plan_name',
		t.content_name,
		'$.materials',
	JSON_ARRAY());

-- insert live_lesson_material
INSERT INTO schedules ( id, tmp_content_id, tmp_content_name ) SELECT
* 
FROM
	(
	SELECT
		assessments.schedule_id AS id,
		assessments_contents.content_id AS content_id,
		cms_contents.content_name AS content_name 
	FROM
		assessments_contents
		INNER JOIN assessments ON assessments_contents.assessment_id = assessments.id
		LEFT JOIN cms_contents ON assessments_contents.content_id = cms_contents.id 
	WHERE
		assessments_contents.content_type = 1 
	) AS t 
	ON DUPLICATE KEY UPDATE live_lesson_plan = JSON_ARRAY_APPEND(
		live_lesson_plan,
		'$.materials',
	JSON_SET( "{}", '$.lesson_material_id', t.content_id, '$.lesson_material_name', t.content_name ));

-- drop tmp columns
ALTER TABLE schedules DROP tmp_content_id;
ALTER TABLE schedules DROP tmp_content_name;