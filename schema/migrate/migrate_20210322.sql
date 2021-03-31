CREATE TABLE IF NOT EXISTS `sets` (
    `id` varchar(50) NOT NULL COMMENT  'id',
    `name` varchar(100) NOT NULL DEFAULT "" COMMENT  'name',
    `organization_id` VARCHAR(50) NOT NULL  DEFAULT "" COMMENT 'organization_id',
    `create_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'created_at',
    `update_at` BIGINT NOT NULL DEFAULT 0  COMMENT 'updated_at',
    `delete_at` BIGINT DEFAULT NULL COMMENT 'deleted_at',
    PRIMARY KEY (`id`),
    UNIQUE  KEY `name_organization_id` (`name`, `organization_id`),
    FULLTEXT KEY `fullindex_name` (`name`)
) COMMENT 'sets' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `outcomes_sets` (
    `id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT  'id',
    `outcome_id` varchar(50) NOT NULL DEFAULT "" COMMENT  'outcome_id',
    `set_id` varchar(50) NOT NULL DEFAULT "" COMMENT  'set_id',
    `create_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'created_at',
    `update_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'updated_at',
    `delete_at` BIGINT DEFAULT NULL COMMENT 'deleted_at',
    UNIQUE KEY `outcome_set_id_delete` (`outcome_id`, `set_id`, `delete_at`)
) COMMENT 'outcomes_sets' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;


alter table learning_outcomes add fulltext fullindex_name (name);
alter table learning_outcomes add fulltext fullindex_keywords(keywords);
alter table learning_outcomes add fulltext fullindex_description(description);