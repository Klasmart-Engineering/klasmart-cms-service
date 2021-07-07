-- assessments
ALTER TABLE assessments `type` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'DEPRECATED: assessment type';

-- assessments_outcomes
ALTER TABLE assessments_outcomes `skip` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'skip';
ALTER TABLE assessments_outcomes `none_achieved` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'none achieved';
