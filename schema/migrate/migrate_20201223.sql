drop index fullindex_name_description_keywords_author_shortcode on learning_outcomes;
create fulltext index fullindex_name_description_keywords_author_shortcode on learning_outcomes(
    `name`,
    `keywords`,
    `description`,
    `author_name`,
    `shortcode`
);

drop index fullindex_name_description_keywords_author_shortcode on learning_outcomes;
alter table learning_outcomes add fulltext index fullindex_name_description_keywords_shortcode(`name`, `keywords`, `description`, `shortcode`);
