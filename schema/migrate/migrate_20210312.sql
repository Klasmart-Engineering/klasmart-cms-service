ALTER TABLE `organizations_properties` ADD COLUMN region varchar(100) NOT NULL DEFAULT '' COMMENT 'region';

insert into organizations_properties(id, type, region) values('9d42af2a-d943-4bb7-84d8-9e2e28b0e290', 'headquarters', 'vn');
update organizations_properties set region = 'global' where id = '10f38ce9-5152-4049-b4e7-6d2e2ba884e6';
