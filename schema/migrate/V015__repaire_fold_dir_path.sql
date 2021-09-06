update
    cms_folder_items,
    (
        select
            cfi1.id as child_id,
            cfi2.id as parent_id,
            cfi1.dir_path as child_dir_path,
            cfi2.dir_path as parent_dir_path
        from
            cms_folder_items cfi1
            left join
            cms_folder_items cfi2
            on cfi1.parent_id=cfi2.id
        where concat(if(cfi2.dir_path='/', '', cfi2.dir_path), '/', cfi2.id) != cfi1.dir_path and cfi1.delete_at=0 and cfi2.delete_at=0 and cfi1.item_type=1
    ) as tm
set
    dir_path=concat(tm.parent_dir_path, if(tm.parent_dir_path='/' or tm.parent_dir_path='', '', '/'), tm.parent_id),
    extra=concat(extra, if(extra is null or extra='', '', ';'), 'repair_folder_0906')
where cms_folder_items.id=tm.child_id;
