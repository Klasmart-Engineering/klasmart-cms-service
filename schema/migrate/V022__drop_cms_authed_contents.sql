drop table cms_authed_contents;
alter table cms_contents add index index_dir_path (dir_path);