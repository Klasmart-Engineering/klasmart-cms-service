ALTER TABLE `cms_folder_items` ADD COLUMN `keywords` TEXT NULL COMMENT '关键字';
ALTER TABLE `cms_folder_items` ADD COLUMN `description` TEXT NULL COMMENT '描述';
ALTER TABLE `cms_folder_items` ADD FULLTEXT INDEX `folder_name_description_keywords_author_index` (`name`, `keywords`, `description`);

INSERT INTO cms_content_visibility_settings (`content_id`, `visibility_setting`) SELECT `id`, `publish_scope` FROM cms_contents;