DELETE 
FROM
	milestones_outcomes 
WHERE
	milestone_id IN ( SELECT id FROM milestones WHERE type = 'general' );

DELETE 
FROM
	milestones 
WHERE
	type = 'general';