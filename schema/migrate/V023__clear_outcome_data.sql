-- Migrate subcategory of learning outcome
INSERT INTO outcomes_relations ( master_id, relation_id, relation_type, master_type, create_at, update_at, delete_at ) SELECT
learning_outcomes.id AS master_id,
SUBSTRING_INDEX( SUBSTRING_INDEX( learning_outcomes.skills, ',', numbers.n ), ',', - 1 ) AS relation_id,
"subcategory" AS relation_type,
"outcome" AS master_type,
learning_outcomes.create_at AS create_at,
learning_outcomes.update_at AS update_at,
IFNULL( learning_outcomes.delete_at, 0 ) AS delete_at 
FROM
	(
	SELECT
		1 AS n UNION ALL
	SELECT
		2 UNION ALL
	SELECT
		3 UNION ALL
	SELECT
		4 UNION ALL
	SELECT
		5 UNION ALL
	SELECT
		6 UNION ALL
	SELECT
		7 UNION ALL
	SELECT
		8 UNION ALL
	SELECT
		9 UNION ALL
	SELECT
		10 UNION ALL
	SELECT
		11 UNION ALL
	SELECT
		12 UNION ALL
	SELECT
		13 UNION ALL
	SELECT
		14 UNION ALL
	SELECT
		15 UNION ALL
	SELECT
		16 UNION ALL
	SELECT
		17 UNION ALL
	SELECT
		18 UNION ALL
	SELECT
		19 UNION ALL
	SELECT
		20 
	) numbers
	INNER JOIN learning_outcomes ON LENGTH( learning_outcomes.skills ) > 0 
	AND LENGTH( learning_outcomes.skills ) - LENGTH(
	REPLACE ( learning_outcomes.skills, ',', '' ))>= numbers.n - 1 
GROUP BY
	id,
	relation_id;

-- Migrate grade of learning outcome
INSERT INTO outcomes_relations ( master_id, relation_id, relation_type, master_type, create_at, update_at, delete_at ) SELECT
learning_outcomes.id AS master_id,
SUBSTRING_INDEX( SUBSTRING_INDEX( learning_outcomes.grade, ',', numbers.n ), ',', - 1 ) AS relation_id,
"grade" AS relation_type,
"outcome" AS master_type,
learning_outcomes.create_at AS create_at,
learning_outcomes.update_at AS update_at,
IFNULL( learning_outcomes.delete_at, 0 ) AS delete_at 
FROM
	(
	SELECT
		1 AS n UNION ALL
	SELECT
		2 UNION ALL
	SELECT
		3 UNION ALL
	SELECT
		4 UNION ALL
	SELECT
		5 UNION ALL
	SELECT
		6 UNION ALL
	SELECT
		7 UNION ALL
	SELECT
		8 UNION ALL
	SELECT
		9 UNION ALL
	SELECT
		10 UNION ALL
	SELECT
		11 UNION ALL
	SELECT
		12 UNION ALL
	SELECT
		13 UNION ALL
	SELECT
		14 UNION ALL
	SELECT
		15 UNION ALL
	SELECT
		16 UNION ALL
	SELECT
		17 UNION ALL
	SELECT
		18 UNION ALL
	SELECT
		19 UNION ALL
	SELECT
		20 
	) numbers
	INNER JOIN learning_outcomes ON LENGTH( learning_outcomes.grade ) > 0 
	AND LENGTH( learning_outcomes.grade ) - LENGTH(
	REPLACE ( learning_outcomes.grade, ',', '' ))>= numbers.n - 1 
GROUP BY
	id,
	relation_id;

-- Migrate age of learning outcome
INSERT INTO outcomes_relations ( master_id, relation_id, relation_type, master_type, create_at, update_at, delete_at ) SELECT
learning_outcomes.id AS master_id,
SUBSTRING_INDEX( SUBSTRING_INDEX( learning_outcomes.age, ',', numbers.n ), ',', - 1 ) AS relation_id,
"age" AS relation_type,
"outcome" AS master_type,
learning_outcomes.create_at AS create_at,
learning_outcomes.update_at AS update_at,
IFNULL( learning_outcomes.delete_at, 0 ) AS delete_at 
FROM
	(
	SELECT
		1 AS n UNION ALL
	SELECT
		2 UNION ALL
	SELECT
		3 UNION ALL
	SELECT
		4 UNION ALL
	SELECT
		5 UNION ALL
	SELECT
		6 UNION ALL
	SELECT
		7 UNION ALL
	SELECT
		8 UNION ALL
	SELECT
		9 UNION ALL
	SELECT
		10 UNION ALL
	SELECT
		11 UNION ALL
	SELECT
		12 UNION ALL
	SELECT
		13 UNION ALL
	SELECT
		14 UNION ALL
	SELECT
		15 UNION ALL
	SELECT
		16 UNION ALL
	SELECT
		17 UNION ALL
	SELECT
		18 UNION ALL
	SELECT
		19 UNION ALL
	SELECT
		20 
	) numbers
	INNER JOIN learning_outcomes ON LENGTH( learning_outcomes.age ) > 0 
	AND LENGTH( learning_outcomes.age ) - LENGTH(
	REPLACE ( learning_outcomes.age, ',', '' ))>= numbers.n - 1 
GROUP BY
	id,
	relation_id;


-- Add unique constraint and remove deduplicates
UPDATE outcomes_relations 
SET delete_at = 0 
WHERE
	delete_at IS NULL;

ALTER TABLE outcomes_relations MODIFY delete_at BIGINT NOT NULL DEFAULT 0;

DROP TABLE
IF
	EXISTS tmp_outcomes_relations;
CREATE TABLE tmp_outcomes_relations SELECT
* 
FROM
	outcomes_relations;
TRUNCATE TABLE outcomes_relations;
ALTER TABLE outcomes_relations ADD UNIQUE INDEX uk_master_id_relation_id_relation_type_delete_at ( master_id, relation_id, relation_type, delete_at );
INSERT IGNORE INTO outcomes_relations SELECT
* 
FROM
	tmp_outcomes_relations;
DROP TABLE tmp_outcomes_relations;