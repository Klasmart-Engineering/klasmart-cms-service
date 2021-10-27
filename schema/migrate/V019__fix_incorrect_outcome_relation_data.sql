-- copy outcomes_relations data from milestones_relations
INSERT INTO outcomes_relations ( master_id, master_type, relation_id, relation_type, create_at, update_at, delete_at ) SELECT
master_id,
master_type,
relation_id,
relation_type,
create_at,
update_at,
delete_at 
FROM
	milestones_relations 
WHERE
	master_type = 'outcome';

-- delete outcomes_relations data from milestones_relations
DELETE 
FROM
	milestones_relations 
WHERE
	master_type = 'outcome';