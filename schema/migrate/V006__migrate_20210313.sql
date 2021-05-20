CREATE TABLE IF NOT EXISTS `programs_groups` (
    `program_id` varchar(100) NOT NULL COMMENT  'program id',
    `group_name` varchar(100) NOT NULL DEFAULT "" COMMENT  'group name',
    PRIMARY KEY (`program_id`)
) COMMENT 'programs groups' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

insert into programs_groups(program_id, group_name) values
("f6617737-5022-478d-9672-0354667e0338", "BadaESL"),
("4591423a-2619-4ef8-a900-f5d924939d02", "BadaSTEAM"),
("d1bbdcc5-0d80-46b0-b98e-162e7439058f", "BadaSTEAM"),
("b39edb9a-ab91-4245-94a4-eb2b5007c033", "BadaESL"),
("7a8c5021-142b-44b1-b60b-275c29d132fe", "BadaESL"),
("56e24fa0-e139-4c80-b365-61c9bc42cd3f", "BadaESL"),
("93f293e8-2c6a-47ad-bc46-1554caac99e4", "BadaESL");

-- "f6617737-5022-478d-9672-0354667e0338", // Bada Talk    BadaESL
-- "4591423a-2619-4ef8-a900-f5d924939d02", // Bada Math    BadaSTEAM
-- "d1bbdcc5-0d80-46b0-b98e-162e7439058f", // Bada STEM    BadaSTEAM
-- "b39edb9a-ab91-4245-94a4-eb2b5007c033", // Bada Genius  BadaESL
-- "7a8c5021-142b-44b1-b60b-275c29d132fe", // Bada Read    BadaESL
-- "56e24fa0-e139-4c80-b365-61c9bc42cd3f", // Bada Sound   BadaESL
-- "93f293e8-2c6a-47ad-bc46-1554caac99e4", // Bada Rhyme   BadaESL

-- insert into organizations_properties(id, type, region) values('6cac91e6-0ef2-4be6-9df7-6ea77f7c1928', 'headquarters', 'vn');
-- insert into organizations_properties(id, type, region) values('9d42af2a-d943-4bb7-84d8-9e2e28b0e290', 'headquarters', 'vn');

CREATE TABLE `organizations_regions` (
  `id` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'id',
  `headquarter` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '0' COMMENT 'headquarter',
  `organization_id` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '0' COMMENT 'organization_id',
  `create_at` bigint(20) DEFAULT '0' COMMENT 'created_at',
  `update_at` bigint(20) DEFAULT '0' COMMENT 'updated_at',
  `delete_at` bigint(20) DEFAULT '0' COMMENT 'delete_at',
  PRIMARY KEY (`id`),
  KEY `organization_regions_headquarter_index` (`headquarter`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='organization_regions';

-- INSERT INTO `organizations_regions` (`id`, `headquarter`, `organization_id`, `create_at`, `update_at`) VALUES
-- ('5fb24528993e7591084c2c18', '6cac91e6-0ef2-4be6-9df7-6ea77f7c1928', 'c0ecdf39-4e20-4f68-88e3-20df10af8b94', 1615963300, 1615963300);
-- Business sql for VN HQ
INSERT INTO `organizations_regions` (`id`, `headquarter`, `organization_id`, `create_at`, `update_at`) VALUES
('5fb24528993e7591084c2c46', '9d42af2a-d943-4bb7-84d8-9e2e28b0e290', '281e49c6-a1f8-4d5e-83f2-0cf76700601c', 1615963415, 1615963415);

