-- cms_contents
ALTER TABLE cms_contents MODIFY 
 `program` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'program';
 
ALTER TABLE cms_contents MODIFY 
 `subject` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'subject';
 
ALTER TABLE cms_contents MODIFY 
 `developmental` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'developmental';
 
 ALTER TABLE cms_contents MODIFY 
 `skills` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'skills';
 
 ALTER TABLE cms_contents MODIFY 
 `age` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'age';
 
 ALTER TABLE cms_contents MODIFY 
 `grade` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'grade';
 
 
-- learning_outcomes
 ALTER TABLE learning_outcomes MODIFY 
 `program` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'program';
 
ALTER TABLE learning_outcomes MODIFY 
 `subject` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'subject';
 
ALTER TABLE learning_outcomes MODIFY 
 `developmental` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'developmental';
 
 ALTER TABLE learning_outcomes MODIFY 
 `skills` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'skills';
 
 ALTER TABLE learning_outcomes MODIFY 
 `age` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'age';
 
 ALTER TABLE learning_outcomes MODIFY 
 `grade` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'grade';