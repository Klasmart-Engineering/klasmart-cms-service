-- remove Task class_type
DELETE 
FROM
	class_types 
WHERE
	id = 'Task';

-- soft delete exist Task schedule data (excluding relation)
SET @cur = UNIX_TIMESTAMP();
UPDATE schedules 
SET delete_at = @cur 
WHERE
	class_type = 'Task' 
	AND delete_at = 0;