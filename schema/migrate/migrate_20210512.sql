alter table milestones add type varchar(10) default 'normal' comment 'milestone type';
alter table milestones modify column name varchar(200) default '';