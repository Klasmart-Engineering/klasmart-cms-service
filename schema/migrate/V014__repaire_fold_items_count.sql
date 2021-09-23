update
    cms_folder_items,
    (
        select cfi.id, items_count, ifnull(c.cc, 0) cc, ifnull(f.fc, 0) fc, ifnull(c.cc, 0)+ifnull(f.fc, 0) as total
        from cms_folder_items cfi
                 left join (
            select SUBSTRING_INDEX(dir_path, '/', -1) as parent_id, count(0) as cc from cms_contents where SUBSTRING_INDEX(dir_path, '/', -1) != '' and publish_status='published' and (delete_at=0 or  delete_at is null)
            group by SUBSTRING_INDEX(dir_path, '/', -1)
        ) c on c.parent_id = cfi.id
                 left join (
            select parent_id, count(0) as fc from cms_folder_items where parent_id != '/' and (delete_at = 0 or delete_at is null) and item_type=1
            group by parent_id
        ) f on f.parent_id = cfi.id
        where 1=1
        and ifnull(c.cc, 0)+ifnull(f.fc, 0) != cfi.items_count
        and cfi.item_type=1
        and (cfi.delete_at=0 or cfi.delete_at is null)
    ) as temp
set cms_folder_items.items_count=temp.total,
    cms_folder_items.extra=concat(if(cms_folder_items.extra is null or cms_folder_items.extra='', '', ';'), 'repair_folder_0902')
where cms_folder_items.id=temp.id;
