ALTER TABLE cms_folder_items MODIFY COLUMN dir_path varchar(768) NOT NULL DEFAULT '/' COMMENT 'Directory path';
ALTER TABLE cms_contents MODIFY COLUMN dir_path varchar(768) NOT NULL DEFAULT '/' COMMENT 'Directory path';

CREATE INDEX idx_cms_contents_dir_path ON cms_contents(dir_path);
CREATE INDEX idx_cms_folder_items_dir_path ON cms_folder_items(dir_path);