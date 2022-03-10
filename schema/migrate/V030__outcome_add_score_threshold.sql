-- add score_threshold field
alter table learning_outcomes add score_threshold float(3,2) NOT NULL DEFAULT 0 COMMENT 'score threshold for auto assessment';

-- set score_threshold default value
update learning_outcomes set score_threshold = 0.8 where assumed = 0;