alter table cms_folder_items add has_descendant tinyint(1) NOT NULL DEFAULT 0 COMMENT 'has published descendant';

create table has_descendant_folder as (
    select distinct substring_index(substring_index(dir_path, '/', b.id+1), '/', -1) folder_id
    from cms_contents a join (select 1 as id union all select 2 union all select 3 union all select 4 union all select 5
                              union all
                              select 6 union all select 7 union all select 8 union all select 9 union all select 10
                              union all
                              select 11 union all select 12 union all select 13 union all select 14 union all select 15) b on b.id < (length(a.dir_path)-length(replace(a.dir_path, '/', ''))+1)
    where a.publish_status='published' and a.delete_at=0 and a.dir_path<>'/');

update cms_folder_items a, has_descendant_folder b  set has_descendant=1 where a.id=b.folder_id and a.delete_at=0;

drop table  if exists has_descendant_folder;