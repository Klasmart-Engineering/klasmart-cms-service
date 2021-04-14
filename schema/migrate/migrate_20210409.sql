CREATE TABLE IF NOT EXISTS `milestones` (
    `id` varchar(50) NOT NULL COMMENT  'id',
    `name` varchar(100) NOT NULL DEFAULT "" COMMENT  'name',
    `shortcode`  CHAR(20) NOT NULL COMMENT 'shortcode',
    `organization_id` varchar(50) NOT NULL COMMENT  'org id',
    `author_id` varchar(50) NOT NULL COMMENT  'author id',
    `status` VARCHAR(16) NOT NULL COMMENT 'status',
    `describe` TEXT NULL COMMENT 'description',
    `ancestor_id` VARCHAR(50) NOT NULL COMMENT 'ancestor',
    `locked_by` VARCHAR(50) COMMENT 'who is editing',
    `source_id` VARCHAR(50) COMMENT 'previous version',
    `latest_id` VARCHAR(50) COMMENT 'latest version',
    `lo_counts` INT NOT NULL DEFAULT  0 COMMENT 'learning outcome counts',
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
    `outcome_id` varchar(50) NOT NULL DEFAULT "" COMMENT  'outcome',
    `create_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'created_at',
    `update_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'updated_at',
    `delete_at` BIGINT DEFAULT NULL COMMENT 'deleted_at',
    UNIQUE KEY `milestone_outcome_id_delete` (`milestone_id`, `outcome_id`, `delete_at`)
) COMMENT 'milestones_outcomes' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE IF NOT EXISTS `milestones_attaches` (
    `id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT  'id',
    `master_id` varchar(50) NOT NULL DEFAULT "" COMMENT  'master resource',
    `attach_id` varchar(50) NOT NULL DEFAULT "" COMMENT  'attach resource',
    `attach_type` varchar(30) NOT NULL DEFAULT "" COMMENT  'attach type',
    `master_type` varchar(30) DEFAULT NULL COMMENT  'master type',
    `create_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'created_at',
    `update_at` BIGINT NOT NULL DEFAULT 0 COMMENT 'updated_at',
    `delete_at` BIGINT DEFAULT NULL COMMENT 'deleted_at',
    UNIQUE KEY `milestone_outcome_id_delete` (`master_id`, `attach_id`, `attach_type`,`master_type`,`delete_at`)
) COMMENT 'milestones_attaches' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;
