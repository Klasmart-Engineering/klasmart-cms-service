#update cms_contents table, add self_study, draw_activity and lesson_type
alter table `cms_contents` add self_study tinyint default 0 not null comment 'is content can self study';
alter table `cms_contents` add draw_activity tinyint default 0 not null comment 'is activity can draw';
alter table `cms_contents` add lesson_type tinyint default 0 not null comment 'lesson type';
alter table `cms_contents` add remark varchar(255) comment 'reject reason remark';
alter table `cms_contents` add source_type varchar(255) comment 'content source type';