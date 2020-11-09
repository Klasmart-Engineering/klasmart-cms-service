CREATE TABLE `ages` (
`id` varchar(256) NOT NULL COMMENT  'id',
`name` varchar(255) DEFAULT NULL COMMENT  'name',
PRIMARY KEY (`id`)
) COMMENT 'ages' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE `lesson_types` (
`id` varchar(256) NOT NULL COMMENT  'id',
`name` varchar(255) DEFAULT NULL COMMENT  'name',
PRIMARY KEY (`id`)
) COMMENT 'lesson_types' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE `grades` (
`id` varchar(256) NOT NULL COMMENT  'id',
`name` varchar(255) DEFAULT NULL COMMENT  'name',
PRIMARY KEY (`id`)
) COMMENT 'grades' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE `developmentals` (
`id` varchar(256) NOT NULL COMMENT  'id',
`name` varchar(255) DEFAULT NULL COMMENT  'name',
PRIMARY KEY (`id`)
) COMMENT 'developmentals' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE `class_types` (
`id` varchar(256) NOT NULL COMMENT  'id',
`name` varchar(255) DEFAULT NULL COMMENT  'name',
PRIMARY KEY (`id`)
) COMMENT 'class_types' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE `programs` (
`id` varchar(256) NOT NULL COMMENT  'id',
`name` varchar(255) DEFAULT NULL COMMENT  'name',
PRIMARY KEY (`id`)
) COMMENT 'programs' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE `subjects` (
`id` varchar(256) NOT NULL COMMENT  'id',
`name` varchar(255) DEFAULT NULL COMMENT  'name',
PRIMARY KEY (`id`)
) COMMENT 'subjects' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE `skills` (
`id` varchar(256) NOT NULL COMMENT  'id',
`name` varchar(255) DEFAULT NULL COMMENT  'name',
PRIMARY KEY (`id`)
) COMMENT 'skills' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE `visibility_settings` (
`id` varchar(256) NOT NULL COMMENT  'id',
`name` varchar(255) DEFAULT NULL COMMENT  'name',
PRIMARY KEY (`id`)
) COMMENT 'visibility_settings' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE IF NOT EXISTS `programs_ages` (
  `id` varchar(50) NOT NULL COMMENT 'id',
  `program_id` varchar(100) NOT NULL COMMENT 'program_id',
  `age_id` varchar(100) NOT NULL COMMENT 'age_id',
  PRIMARY KEY (`id`),
  KEY `idx_program_id` (`program_id`),
  KEY `idx_age_id` (`age_id`)
) ENGINE=InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT='programs_ages';

CREATE TABLE IF NOT EXISTS `programs_developments` (
  `id` varchar(50) NOT NULL COMMENT 'id',
  `program_id` varchar(100) NOT NULL COMMENT 'program_id',
  `development_id` varchar(100) NOT NULL COMMENT 'development_id',
  PRIMARY KEY (`id`),
  KEY `idx_program_id` (`program_id`),
  KEY `idx_development_id` (`development_id`)
) ENGINE=InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT='programs_developments';

CREATE TABLE IF NOT EXISTS `programs_grades` (
  `id` varchar(50) NOT NULL COMMENT 'id',
  `program_id` varchar(100) NOT NULL COMMENT 'program_id',
  `grade_id` varchar(100) NOT NULL COMMENT 'grade_id',
  PRIMARY KEY (`id`),
  KEY `idx_program_id` (`program_id`),
  KEY `idx_grade_id` (`grade_id`)
) ENGINE=InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT='programs_grades';

CREATE TABLE IF NOT EXISTS `developments_skills` (
  `id` varchar(50) NOT NULL COMMENT 'id',
  `program_id` varchar(100) NOT NULL COMMENT 'program_id',
  `development_id` varchar(100) NOT NULL COMMENT 'development_id',
  `skill_id` varchar(100) NOT NULL COMMENT 'skill_id',
  PRIMARY KEY (`id`),
  KEY `idx_program_id` (`program_id`),
  KEY `idx_development_id` (`development_id`),
  KEY `idx_skill_id` (`skill_id`),
  Key `idx_program_develop_skill` (`program_id`,`development_id`,`skill_id`)
) ENGINE=InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT='developments_skills';

CREATE TABLE IF NOT EXISTS `programs_subjects` (
  `id` varchar(50) NOT NULL COMMENT 'id',
  `program_id` varchar(100) NOT NULL COMMENT 'program_id',
  `subject_id` varchar(100) NOT NULL COMMENT 'subject_id',
  PRIMARY KEY (`id`),
  KEY `idx_program_id` (`program_id`),
  KEY `idx_subject_id` (`subject_id`)
) ENGINE=InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT='programs_subjects';

-- init data
INSERT INTO `ages` (`id`,`name`) VALUES ("age1","3-4"),("age2","4-5"),("age3","5-6"),("age4","6-7"),("age5","7-8");
INSERT INTO `class_types` (`id`,`name`) VALUES ("OnlineClass","schedule_detail_online_class"),("OfflineClass","schedule_detail_offline_class"),("Homework","schedule_detail_homework"),("Task","schedule_detail_task");
INSERT INTO `developmentals` (`id`,`name`) VALUES ("developmental1","Speech & Language Skills"),("developmental2","Fine Motor Skills"),("developmental3","Gross Motor Skills"),("developmental4","Cognitive Skills"),("developmental5","Personal Development"),("developmental6","Language and Numeracy Skills"),("developmental7","Cognitive"),("developmental8","Social and Emotional"),("developmental9","Oral"),("developmental10","Literacy"),("developmental11","Whole-Child"),("developmental12","Knowledge");
INSERT INTO `grades` (`id`,`name`) VALUES ("grade1","Not Specific"),("grade2","PreK-1"),("grade3","PreK-2"),("grade4","K"),("grade5","Grade 1"),("grade6","Grade 2"),("grade12","Kindergarten"),("grade7","PreK-3"),("grade8","PreK-4"),("grade9","PreK-5"),("grade10","PreK-6"),("grade11","PreK-7");
INSERT INTO `lesson_types` (`id`,`name`) VALUES ("1","Test"),("2","Not Test");
INSERT INTO `programs` (`id`,`name`) VALUES ("program1","Badanamu ESL"),("program2","Bada Math"),("program3","Bada STEM"),("program4","ESL"),("program5","Math"),("program6","Science"),("program7","Bada Rhyme");
INSERT INTO `skills` (`id`,`name`) VALUES ("skills1","Speaking & Listening"),("skills2","Phonics"),("skills3","Vocabulary"),("skills4","Thematic Concepts"),("skills5","Reading Skills and Comprehension"),("skills6","Sight Words"),("skills7","Sensory"),("skills8","Hand-Eye Coordination"),("skills9","Simple Movements"),("skills10","Complex Movements"),("skills11","Physical Skills"),("skills12","Logic & Memory"),("skills13","Visual"),("skills14","Social Skills"),("skills15","Emotional Skills"),("skills16","Reasoning"),("skills17","Listening"),("skills18","Speaking"),("skills19","Interpreting"),("skills20","Numbers"),("skills21","Fluency"),("skills25","Spatial Representation"),("skills26","Counting and Operations"),("skills27","Logical Problem-Solving"),("skills28","Patterns"),("skills29","Social Interactions"),("skills30","Empathy"),("skills31","Self-Identity"),("skills32","Self-Control"),("skills44","Phonological Awareness"),("skills45","Language Support"),("skills46","Communication"),("skills47","Emergent Reading"),("skills48","Emergent Writing"),("skills49","Social-Emotional Learning"),("skills50","Cognitive Development"),("skills51","Physical Coordination"),("skills40","Science"),("skills52","Technology"),("skills41","Math"),("skills53","Engineering"),("skills54","Miscellaneous");
INSERT INTO `subjects` (`id`,`name`) VALUES ("subject1","Language/Literacy"),("subject2","Math"),("subject3","Science");
INSERT INTO `visibility_settings` (`id`,`name`) VALUES ("visibility_settings1","library_label_visibility_school"),("visibility_settings2","library_label_visibility_organization");

INSERT INTO `programs_ages` (`id`,`program_id`,`age_id`) VALUES ("5f9fce99f3f631066bfc58e0","program1","age1"),("5f9fce99f3f631066bfc58e1","program1","age2"),("5f9fce99f3f631066bfc58e2","program1","age3"),("5f9fce99f3f631066bfc58e3","program1","age4"),("5f9fce99f3f631066bfc58e4","program1","age5"),("5f9fce99f3f631066bfc58e5","program2","age1"),("5f9fce99f3f631066bfc58e6","program2","age2"),("5f9fce99f3f631066bfc58e7","program2","age3"),("5f9fce99f3f631066bfc58e8","program2","age4"),("5f9fce99f3f631066bfc58e9","program2","age5"),("5f9fce99f3f631066bfc58ea","program3","age1"),("5f9fce99f3f631066bfc58eb","program3","age2"),("5f9fce99f3f631066bfc58ec","program3","age3"),("5f9fce99f3f631066bfc58ed","program3","age4"),("5f9fce99f3f631066bfc58ee","program3","age5"),("5f9fce99f3f631066bfc58ef","program4","age1"),("5f9fce99f3f631066bfc58f0","program4","age2"),("5f9fce99f3f631066bfc58f1","program4","age3"),("5f9fce99f3f631066bfc58f2","program4","age4"),("5f9fce99f3f631066bfc58f3","program5","age1"),("5f9fce99f3f631066bfc58f4","program5","age2"),("5f9fce99f3f631066bfc58f5","program5","age3"),("5f9fce99f3f631066bfc58f6","program5","age4"),("5f9fce99f3f631066bfc58f7","program5","age5"),("5f9fce99f3f631066bfc58f8","program6","age1"),("5f9fce99f3f631066bfc58f9","program6","age2"),("5f9fce99f3f631066bfc58fa","program6","age3"),("5f9fce99f3f631066bfc58fb","program6","age4"),("5f9fce99f3f631066bfc58fc","program6","age5"),("5f9fce99f3f631066bfc58fd","program7","age1"),("5f9fce99f3f631066bfc58fe","program7","age2"),("5f9fce99f3f631066bfc58ff","program7","age3");
INSERT INTO `programs_developments` (`id`,`program_id`,`development_id`) VALUES ("5f9fce99f3f631066bfc5900","program1","developmental1"),("5f9fce99f3f631066bfc5901","program1","developmental2"),("5f9fce99f3f631066bfc5902","program1","developmental3"),("5f9fce99f3f631066bfc5903","program1","developmental4"),("5f9fce99f3f631066bfc5904","program1","developmental5"),("5f9fce99f3f631066bfc5905","program2","developmental6"),("5f9fce99f3f631066bfc5906","program2","developmental2"),("5f9fce99f3f631066bfc5907","program2","developmental3"),("5f9fce99f3f631066bfc5908","program2","developmental7"),("5f9fce99f3f631066bfc5909","program2","developmental8"),("5f9fce99f3f631066bfc590a","program3","developmental1"),("5f9fce99f3f631066bfc590b","program3","developmental2"),("5f9fce99f3f631066bfc590c","program3","developmental3"),("5f9fce99f3f631066bfc590d","program3","developmental4"),("5f9fce99f3f631066bfc590e","program3","developmental8"),("5f9fce99f3f631066bfc590f","program4","developmental1"),("5f9fce99f3f631066bfc5910","program4","developmental2"),("5f9fce99f3f631066bfc5911","program4","developmental3"),("5f9fce99f3f631066bfc5912","program4","developmental4"),("5f9fce99f3f631066bfc5913","program4","developmental5"),("5f9fce99f3f631066bfc5914","program5","developmental6"),("5f9fce99f3f631066bfc5915","program5","developmental2"),("5f9fce99f3f631066bfc5916","program5","developmental3"),("5f9fce99f3f631066bfc5917","program5","developmental4"),("5f9fce99f3f631066bfc5918","program5","developmental5"),("5f9fce99f3f631066bfc5919","program6","developmental6"),("5f9fce99f3f631066bfc591a","program6","developmental2"),("5f9fce99f3f631066bfc591b","program6","developmental3"),("5f9fce99f3f631066bfc591c","program6","developmental4"),("5f9fce99f3f631066bfc591d","program6","developmental5"),("5f9fce99f3f631066bfc591e","program7","developmental6"),("5f9fce99f3f631066bfc591f","program7","developmental2"),("5f9fce99f3f631066bfc5920","program7","developmental3"),("5f9fce99f3f631066bfc5921","program7","developmental4"),("5f9fce99f3f631066bfc5922","program7","developmental5"),("5f9fce99f3f631066bfc5923","program7","developmental9"),("5f9fce99f3f631066bfc5924","program7","developmental10"),("5f9fce99f3f631066bfc5925","program7","developmental11"),("5f9fce99f3f631066bfc5926","program7","developmental12");
INSERT INTO `programs_grades` (`id`,`program_id`,`grade_id`) VALUES ("5f9fce99f3f631066bfc5927","program1","grade1"),("5f9fce99f3f631066bfc5928","program2","grade2"),("5f9fce99f3f631066bfc5929","program2","grade3"),("5f9fce99f3f631066bfc592a","program2","grade4"),("5f9fce99f3f631066bfc592b","program2","grade5"),("5f9fce99f3f631066bfc592c","program2","grade6"),("5f9fce99f3f631066bfc592d","program3","grade2"),("5f9fce99f3f631066bfc592e","program3","grade3"),("5f9fce99f3f631066bfc592f","program3","grade4"),("5f9fce99f3f631066bfc5930","program3","grade5"),("5f9fce99f3f631066bfc5931","program3","grade6"),("5f9fce99f3f631066bfc5932","program4","grade2"),("5f9fce99f3f631066bfc5933","program4","grade3"),("5f9fce99f3f631066bfc5934","program4","grade12"),("5f9fce99f3f631066bfc5935","program4","grade5"),("5f9fce99f3f631066bfc5936","program5","grade7"),("5f9fce99f3f631066bfc5937","program5","grade8"),("5f9fce99f3f631066bfc5938","program5","grade9"),("5f9fce99f3f631066bfc5939","program5","grade10"),("5f9fce99f3f631066bfc593a","program5","grade11"),("5f9fce99f3f631066bfc593b","program6","grade7"),("5f9fce99f3f631066bfc593c","program6","grade8"),("5f9fce99f3f631066bfc593d","program6","grade9"),("5f9fce99f3f631066bfc593e","program6","grade10"),("5f9fce99f3f631066bfc593f","program6","grade11"),("5f9fce99f3f631066bfc5940","program7","grade1");
INSERT INTO `developments_skills` (`id`,`program_id`,`development_id`,`skill_id`) VALUES ("5fa2583c14afd865cba2003f","program1","developmental1","skills1"),("5fa2583c14afd865cba20040","program1","developmental1","skills2"),("5fa2583c14afd865cba20041","program1","developmental1","skills3"),("5fa2583c14afd865cba20042","program1","developmental1","skills4"),("5fa2583c14afd865cba20043","program1","developmental1","skills5"),("5fa2583c14afd865cba20044","program1","developmental1","skills6"),("5fa2583c14afd865cba20045","program1","developmental2","skills7"),("5fa2583c14afd865cba20046","program1","developmental2","skills8"),("5fa2583c14afd865cba20047","program1","developmental3","skills9"),("5fa2583c14afd865cba20048","program1","developmental3","skills10"),("5fa2583c14afd865cba20049","program1","developmental3","skills11"),("5fa2583c14afd865cba2004a","program1","developmental4","skills12"),("5fa2583c14afd865cba2004b","program1","developmental4","skills13"),("5fa2583c14afd865cba2004c","program1","developmental5","skills14"),("5fa2583c14afd865cba2004d","program1","developmental5","skills15"),("5fa2583c14afd865cba2004e","program2","developmental6","skills16"),("5fa2583c14afd865cba2004f","program2","developmental6","skills3"),("5fa2583c14afd865cba20050","program2","developmental6","skills17"),("5fa2583c14afd865cba20051","program2","developmental6","skills18"),("5fa2583c14afd865cba20052","program2","developmental6","skills19"),("5fa2583c14afd865cba20053","program2","developmental6","skills20"),("5fa2583c14afd865cba20054","program2","developmental6","skills21"),("5fa2583c14afd865cba20055","program2","developmental7","skills25"),("5fa2583c14afd865cba20056","program2","developmental7","skills26"),("5fa2583c14afd865cba20057","program2","developmental7","skills27"),("5fa2583c14afd865cba20058","program2","developmental7","skills28"),("5fa2583c14afd865cba20059","program2","developmental8","skills29"),("5fa2583c14afd865cba2005a","program2","developmental8","skills30"),("5fa2583c14afd865cba2005b","program2","developmental8","skills31"),("5fa2583c14afd865cba2005c","program2","developmental8","skills32"),("5fa2583c14afd865cba2005d","program7","developmental9","skills44"),("5fa2583c14afd865cba2005e","program7","developmental9","skills3"),("5fa2583c14afd865cba2005f","program7","developmental9","skills45"),("5fa2583c14afd865cba20060","program7","developmental9","skills46"),("5fa2583c14afd865cba20061","program7","developmental10","skills47"),("5fa2583c14afd865cba20062","program7","developmental10","skills48"),("5fa2583c14afd865cba20063","program7","developmental11","skills49"),("5fa2583c14afd865cba20064","program7","developmental11","skills50"),("5fa2583c14afd865cba20065","program7","developmental11","skills51"),("5fa2583c14afd865cba20066","program7","developmental12","skills40"),("5fa2583c14afd865cba20067","program7","developmental12","skills52"),("5fa2583c14afd865cba20068","program7","developmental12","skills41"),("5fa2583c14afd865cba20069","program7","developmental12","skills53"),("5fa2583c14afd865cba2006a","program7","developmental12","skills54");
INSERT INTO `programs_subjects` (`id`,`program_id`,`subject_id`) VALUES ("5f9fce99f3f631066bfc596d","program1","subject1"),("5f9fce99f3f631066bfc596e","program2","subject2"),("5f9fce99f3f631066bfc596f","program3","subject3"),("5f9fce99f3f631066bfc5970","program4","subject1"),("5f9fce99f3f631066bfc5971","program5","subject2"),("5f9fce99f3f631066bfc5972","program6","subject3"),("5f9fce99f3f631066bfc5973","program7","subject1");

INSERT INTO `ages` (`id`,`name`) VALUES ("age0","None Specified");
INSERT INTO `developmentals` (`id`,`name`) VALUES ("developmental0","None Specified");
INSERT INTO `programs` (`id`,`name`) VALUES ("program0","None Specified");
INSERT INTO `skills` (`id`,`name`) VALUES ("skills0","None Specified");
INSERT INTO `subjects` (`id`,`name`) VALUES ("subject0","None Specified");
INSERT INTO `grades` (`id`,`name`) VALUES ("grade0","None Specified");
