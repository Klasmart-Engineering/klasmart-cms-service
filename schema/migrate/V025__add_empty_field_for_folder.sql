alter table cms_folder_items add has_descendant tinyint(1) NOT NULL DEFAULT 0 COMMENT 'has published descendant';

create table has_descendant_folder as (
    select distinct substring_index(substring_index(dir_path, '/', b.id+1), '/', -1) folder_id
    from cms_contents a join (select 1 as id union select 2 as id union select 3 as id union select 4 as id union select 5 as id
                              union
                              select 6 as id union select 7 as id union select 8 as id union select 9 as id union select 10 as id
                              union
                              select 11 as id union select 12 as id union select 13 as id union select 14 as id union select 15 as id) b on b.id < (length(a.dir_path)-length(replace(a.dir_path, '/', ''))+1)
    where a.publish_status='published' and a.delete_at=0 and a.dir_path<>'/');

update cms_folder_items a, has_descendant_folder b  set has_descendant=1 where a.id=b.folder_id and a.delete_at=0;

drop table  if exists has_descendant_folder;