ALTER TABLE `programs` ADD COLUMN group_name varchar(100) DEFAULT NULL COMMENT 'group_name';
ALTER TABLE `programs` ADD INDEX idx_group_name_delete_at ( `group_name`, `delete_at`);