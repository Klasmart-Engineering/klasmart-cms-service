insert into cms_content_properties
(content_id, property_type, property_id, sequence)
    (
        select
            cms_contents.id,
            2,
            SUBSTRING_INDEX(SUBSTRING_INDEX(cms_contents.subject, ',', numbers.n), ',', -1) subject,
            numbers.n - 1
        from
            (select 1 n union all
             select 2 union all select 3 union all
             select 4 union all select 5) numbers INNER JOIN cms_contents
                                                             on CHAR_LENGTH(cms_contents.subject)
                                                                    -CHAR_LENGTH(REPLACE(cms_contents.subject, ',', ''))>=numbers.n-1
    );

insert into cms_content_properties
(content_id, property_type, property_id, sequence)
    (
        select
            cms_contents.id,
            1,
            SUBSTRING_INDEX(SUBSTRING_INDEX(cms_contents.program, ',', numbers.n), ',', -1) program,
            numbers.n - 1
        from
            (select 1 n union all
             select 2 union all select 3 union all
             select 4 union all select 5) numbers INNER JOIN cms_contents
                                                             on CHAR_LENGTH(cms_contents.program)
                                                                    -CHAR_LENGTH(REPLACE(cms_contents.program, ',', ''))>=numbers.n-1
    );

insert into cms_content_properties
(content_id, property_type, property_id, sequence)
    (
        select
            cms_contents.id,
            3,
            SUBSTRING_INDEX(SUBSTRING_INDEX(cms_contents.developmental, ',', numbers.n), ',', -1) developmental,
            numbers.n - 1
        from
            (select 1 n union all
             select 2 union all select 3 union all
             select 4 union all select 5) numbers INNER JOIN cms_contents
                                                             on CHAR_LENGTH(cms_contents.developmental)
                                                                    -CHAR_LENGTH(REPLACE(cms_contents.developmental, ',', ''))>=numbers.n-1
    );

insert into cms_content_properties
(content_id, property_type, property_id, sequence)
    (
        select
            cms_contents.id,
            4,
            SUBSTRING_INDEX(SUBSTRING_INDEX(cms_contents.age, ',', numbers.n), ',', -1) age,
            numbers.n - 1
        from
            (select 1 n union all
             select 2 union all select 3 union all
             select 4 union all select 5) numbers INNER JOIN cms_contents
                                                             on CHAR_LENGTH(cms_contents.age)
                                                                    -CHAR_LENGTH(REPLACE(cms_contents.age, ',', ''))>=numbers.n-1
    );

insert into cms_content_properties
(content_id, property_type, property_id, sequence)
    (
        select
            cms_contents.id,
            5,
            SUBSTRING_INDEX(SUBSTRING_INDEX(cms_contents.grade, ',', numbers.n), ',', -1) grade,
            numbers.n - 1
        from
            (select 1 n union all
             select 2 union all select 3 union all
             select 4 union all select 5) numbers INNER JOIN cms_contents
                                                             on CHAR_LENGTH(cms_contents.grade)
                                                                    -CHAR_LENGTH(REPLACE(cms_contents.grade, ',', ''))>=numbers.n-1
    );


insert into cms_content_properties
(content_id, property_type, property_id, sequence)
    (
        select
            cms_contents.id,
            6,
            SUBSTRING_INDEX(SUBSTRING_INDEX(cms_contents.skills, ',', numbers.n), ',', -1) skills,
            numbers.n - 1
        from
            (select 1 n union all
             select 2 union all select 3 union all
             select 4 union all select 5) numbers INNER JOIN cms_contents
                                                             on CHAR_LENGTH(cms_contents.skills)
                                                                    -CHAR_LENGTH(REPLACE(cms_contents.skills, ',', ''))>=numbers.n-1
    )
;
