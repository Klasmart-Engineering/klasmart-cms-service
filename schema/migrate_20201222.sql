-- add group_name for programs and update data
ALTER TABLE `programs` ADD COLUMN group_name varchar(100) DEFAULT NULL COMMENT 'group_name';
ALTER TABLE `programs` ADD INDEX idx_group_name_delete_at ( `group_name`, `delete_at`);
update programs set group_name = 'BadaESL' where id in ('program4','program5','program7','program6','program1');
update programs set group_name = 'BadaSTEAM' where id in ('program2','program3');