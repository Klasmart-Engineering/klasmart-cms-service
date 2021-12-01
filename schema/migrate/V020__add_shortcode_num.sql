ALTER TABLE learning_outcomes
ADD COLUMN shortcode_num INT NOT NULL DEFAULT 0;

UPDATE learning_outcomes 
SET shortcode = "00000",
shortcode_num = 0 
WHERE
	shortcode = "";
	
UPDATE learning_outcomes 
SET shortcode_num = CAST( CONV( shortcode, 36, 10 ) AS UNSIGNED );

CREATE INDEX index_shortcode_num_organization_id ON learning_outcomes ( organization_id, shortcode_num );

ALTER TABLE milestones
ADD COLUMN shortcode_num INT NOT NULL DEFAULT 0;

UPDATE milestones 
SET shortcode = "00000",
shortcode_num = 0 
WHERE
	shortcode = "";
	
UPDATE milestones 
SET shortcode_num = CAST( CONV( shortcode, 36, 10 ) AS UNSIGNED );

CREATE INDEX index_shortcode_num_organization_id ON milestones ( organization_id, shortcode_num );

