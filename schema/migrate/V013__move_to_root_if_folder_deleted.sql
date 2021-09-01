-- select id, dir_path, delete_at from cms_folder_items where delete_at != 0 and delete_at is not null;
-- select concat(if(dir_path='/','',dir_path), '/', id) as dir_path from cms_folder_items where delete_at != 0 and delete_at is not null;

update cms_contents set dir_path= '/' where dir_path in (
    select concat(if(dir_path='/','', dir_path), '/', id) as dir_path from cms_folder_items where delete_at != 0 and delete_at is not null
);