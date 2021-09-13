alter table cms_contents add parent_folder varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '/' COMMENT 'parent folder id';

update cms_contents
set
    parent_folder=if(dir_path='/' or dir_path='', '/', substring_index(dir_path, '/', -1)),
    extra=concat(extra, if(extra is null or extra='', '', ';'), 'parent_folder_0913');