ALTER TABLE `organizations_properties` ADD COLUMN region varchar(100) NOT NULL DEFAULT '' COMMENT 'region';

insert into organizations_properties(id, type, region) values('6cac91e6-0ef2-4be6-9df7-6ea77f7c1928', 'headquarters', 'vn');
update organizations_properties set region = 'global' where id = '10f38ce9-5152-4049-b4e7-6d2e2ba884e6';
