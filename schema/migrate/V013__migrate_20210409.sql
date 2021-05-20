CREATE TABLE IF NOT EXISTS `milestones` (
    `id` varchar(50) NOT NULL COMMENT  'id',
    `name` varchar(100) NOT NULL DEFAULT "" COMMENT  'name',
    `shortcode`  CHAR(20) NOT NULL DEFAULT "" COMMENT 'shortcode',
    `organization_id` varchar(50) NOT NULL DEFAULT "" COMMENT  'org id',
    `author_id` varchar(50) NOT NULL DEFAULT "" COMMENT  'author id',
    `status` VARCHAR(16) NOT NULL DEFAULT "" COMMENT 'status',
    `describe` TEXT NULL COMMENT 'description',
    `ancestor_id` VARCHAR(50) NOT NULL DEFAULT "" COMMENT 'ancestor',
    `locked_by` VARCHAR(50) DEFAULT "" COMMENT 'who is editing',
    `source_id` VARCHAR(50) DEFAULT "" COMMENT 'previous version',
    `latest_id` VARCHAR(50) DEFAULT "" COMMENT 'latest version',
    `create_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'created_at',
    `update_at` BIGINT NOT NULL DEFAULT 0  COMMENT 'updated_at',
    `delete_at` BIGINT DEFAULT NULL COMMENT 'deleted_at',
    PRIMARY KEY (`id`),
    FULLTEXT KEY `fullindex_name_shortcode_describe` (`name`, `shortcode`, `describe`),
    FULLTEXT KEY `fullindex_name` (`name`),
    FULLTEXT KEY `fullindex_shortcode` (`shortcode`),
    FULLTEXT KEY `fullindex_describe` (`describe`)
) COMMENT 'milestones' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `milestones_outcomes` (
    `id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT  'id',
    `milestone_id` varchar(50) NOT NULL DEFAULT "" COMMENT  'milestone',
    `outcome_ancestor` varchar(50) NOT NULL DEFAULT "" COMMENT  'outcome ancestor',
    `create_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'created_at',
    `update_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'updated_at',
    `delete_at` BIGINT DEFAULT NULL COMMENT 'deleted_at',
    UNIQUE KEY `milestone_ancestor_id_delete` (`milestone_id`, `outcome_ancestor`, `delete_at`)
) COMMENT 'milestones_outcomes' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE IF NOT EXISTS `milestones_relations` (
    `id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT  'id',
    `master_id` varchar(50) NOT NULL DEFAULT "" COMMENT  'master resource',
    `relation_id` varchar(50) NOT NULL DEFAULT "" COMMENT  'relation resource',
    `relation_type` varchar(30) NOT NULL DEFAULT "" COMMENT  'relation type',
    `master_type` varchar(30) DEFAULT NULL COMMENT  'master type',
    `create_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'created_at',
    `update_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'updated_at',
    `delete_at` BIGINT DEFAULT NULL COMMENT 'deleted_at',
    UNIQUE KEY `master_relation_delete` (`master_id`, `relation_id`, `relation_type`,`master_type`,`delete_at`)
) COMMENT 'milestones_relations' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE IF NOT EXISTS `outcomes_relations` (
    `id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT  'id',
    `master_id` varchar(50) NOT NULL DEFAULT "" COMMENT  'master resource',
    `relation_id` varchar(50) NOT NULL DEFAULT "" COMMENT  'relation resource',
    `relation_type` varchar(30) NOT NULL DEFAULT "" COMMENT  'relation type',
    `master_type` varchar(30) DEFAULT NULL COMMENT  'master type',
    `create_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'created_at',
    `update_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'updated_at',
    `delete_at` BIGINT DEFAULT NULL COMMENT 'deleted_at',
    UNIQUE KEY `master_relation_delete` (`master_id`, `relation_id`, `relation_type`,`master_type`,`delete_at`)
) COMMENT 'outcomes_relations' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

alter table milestones drop column type;