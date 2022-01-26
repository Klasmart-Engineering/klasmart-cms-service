alter table cms_folder_items add has_descendant tinyint(1) NOT NULL DEFAULT 0 COMMENT 'has published descendant';

update cms_folder_items set has_descendant = (
        case when exists (select id from cms_contents where position(cms_folder_items.id in dir_path)>0 and publish_status='published' and delete_at=0)
        then 1 else 0 end
);