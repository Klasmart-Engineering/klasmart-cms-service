-- select id, dir_path, delete_at from cms_folder_items where delete_at != 0 and delete_at is not null;
-- select concat(if(dir_path='/','',dir_path), '/', id) as dir_path from cms_folder_items where delete_at != 0 and delete_at is not null;

update cms_contents set dir_path= '/', extra=concat(extra, if(extra is null or extra='', '', ';'), 'repair_folder_0901') where dir_path in (
    select concat(if(dir_path='/','', dir_path), '/', id) as dir_path from cms_folder_items where delete_at != 0 and delete_at is not null
);

delimiter $$

drop procedure if exists repaire_lost_folders$$

create procedure repaire_lost_folders()
begin
    declare wid, wparent_id varchar(50);
    declare wdir_path varchar(768);
    declare dist_path, new_dir_path varchar(768);

    declare done int default false;
    declare cur cursor for select id, parent_id, dir_path from cms_folder_items where dir_path not like concat('%', parent_id);
    declare continue handler for not found set done = true;

    open cur;
        read_loop: loop
            fetch cur into wid, wparent_id, wdir_path;
            if done then
                leave read_loop;
            else
    --             select wid, wparent_id, wdir_path;
                select dir_path into dist_path from cms_folder_items where id=wparent_id;
                set new_dir_path = concat(dist_path, '/', wparent_id);
    --             select new_dir_path;
                update cms_folder_items set dir_path=replace(dir_path, wdir_path, new_dir_path), extra=concat(extra, if(extra is null or extra='', '', ';'), 'repair_folder_0901') where dir_path like concat('%', wid, '%');
                update cms_folder_items set dir_path=new_dir_path, extra=concat(extra, if(extra is null or extra='', '', ';'), 'repair_folder_0901') where id=wid;
            end if;
        end loop;
    close cur;
end$$

delimiter ;

call repaire_lost_folders();
drop procedure if exists repaire_lost_folders;
