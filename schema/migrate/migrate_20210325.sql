ALTER TABLE `cms_folder_items` ADD COLUMN `keywords` TEXT NULL COMMENT '关键字';
ALTER TABLE `cms_folder_items` ADD COLUMN `description` TEXT NULL COMMENT '描述';
ALTER TABLE `cms_folder_items` ADD FULLTEXT INDEX `folder_name_description_keywords_author_index` (`name`, `keywords`, `description`);
