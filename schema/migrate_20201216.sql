-- init headquarters: Badanamu HQ
insert into organizations_properties(id, type) values('0f38ce9-5152-4049-b4e7-6d2e2ba884e6', 'headquarters');

-- add column org_type for programs
ALTER TABLE `programs` ADD COLUMN org_type varchar(100) DEFAULT NULL COMMENT 'org_type';
ALTER TABLE `programs` ADD INDEX idx_org_type_delete_at ( `org_type`, `delete_at`);
update programs set org_type='headquarters' where org_type = ''
