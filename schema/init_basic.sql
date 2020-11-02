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
`developmental_id` varchar(100) DEFAULT NULL COMMENT  'developmental_id',
PRIMARY KEY (`id`)
) COMMENT 'skills' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;

CREATE TABLE `visibility_settings` (
`id` varchar(256) NOT NULL COMMENT  'id',
`name` varchar(255) DEFAULT NULL COMMENT  'name',
PRIMARY KEY (`id`)
) COMMENT 'visibility_settings' DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci ;


-- init data
INSERT INTO `ages` (`id`,`name`) VALUES ("age1","3-4"),("age2","4-5"),("age3","5-6"),("age4","6-7"),("age5","7-8");
INSERT INTO `class_types` (`id`,`name`) VALUES ("OnlineClass","schedule_detail_online_class"),("OfflineClass","schedule_detail_offline_class"),("Homework","schedule_detail_homework"),("Task","schedule_detail_task");
INSERT INTO `developmentals` (`id`,`name`) VALUES ("developmental1","Speech & Language Skills"),("developmental2","Fine Motor Skills"),("developmental3","Gross Motor Skills"),("developmental4","Cognitive Skills"),("developmental5","Personal Development"),("developmental6","Language and Numeracy Skills"),("developmental7","Cognitive"),("developmental8","Social and Emotional"),("developmental9","Oral"),("developmental10","Literacy"),("developmental11","Whole-Child"),("developmental12","Knowledge");
INSERT INTO `grades` (`id`,`name`) VALUES ("grade1","Not Specific"),("grade2","PreK-1"),("grade3","PreK-2"),("grade4","K"),("grade5","Grade 1"),("grade6","Grade 2"),("grade12","Kindergarten"),("grade7","PreK-3"),("grade8","PreK-4"),("grade9","PreK-5"),("grade10","PreK-6"),("grade11","PreK-7");
INSERT INTO `lesson_types` (`id`,`name`) VALUES ("1","Test"),("2","Not Test");
INSERT INTO `programs` (`id`,`name`) VALUES ("program1","Badanamu ESL"),("program2","Bada Math"),("program3","Bada STEM"),("program4","ESL"),("program5","Math"),("program6","Science"),("program7","Bada Rhyme");
INSERT INTO `skills` (`id`,`name`,`developmental_id`) VALUES ("skills1","Speaking & Listening","developmental1"),("skills2","Phonics","developmental1"),("skills3","Vocabulary","developmental1"),("skills4","Thematic Concepts","developmental1"),("skills5","Reading Skills and Comprehension","developmental1"),("skills6","Sight Words","developmental1"),("skills7","Sensory","developmental2"),("skills8","Hand-Eye Coordination","developmental2"),("skills9","Simple Movements","developmental3"),("skills10","Complex Movements","developmental3"),("skills11","Physical Skills","developmental3"),("skills12","Logic & Memory","developmental4"),("skills13","Visual","developmental4"),("skills14","Social Skills","developmental5"),("skills15","Emotional Skills","developmental5"),("skills16","Reasoning","developmental6"),("skills17","Listening","developmental6"),("skills18","Speaking","developmental6"),("skills19","Interpreting","developmental6"),("skills20","Numbers","developmental6"),("skills21","Fluency","developmental6"),("skills22","Academic Skill (Drawing, Tracing, Coloring, Writing)","developmental2"),("skills23","Play Skill (Drag and Drop, Screen Click)","developmental2"),("skills24","Body Coordination","developmental3"),("skills25","Spatial Representation","developmental7"),("skills26","Counting and Operations","developmental7"),("skills27","Logical Problem-Solving","developmental7"),("skills28","Patterns","developmental7"),("skills29","Social Interactions","developmental8"),("skills30","Empathy","developmental8"),("skills31","Self-Identity","developmental8"),("skills32","Self-Control","developmental8"),("skills33","Writing","developmental1"),("skills34","Science Process (Observing, Classifying, Communicating, Measuring, Predicting)","developmental4"),("skills35","Critical Thinking (Interpretation, Analysis, Evaluation, Inference, Explanation, and Self-Regulation)","developmental4"),("skills36","Reasoning Skills","developmental4"),("skills37","Colors","developmental6"),("skills38","Shapes","developmental6"),("skills39","Letters","developmental6"),("skills40","Science","developmental6"),("skills41","Math","developmental6"),("skills42","Coding","developmental6"),("skills43","Experimenting & Problem Solving","developmental4"),("skills44","Phonological Awareness","developmental9"),("skills45","Language Support","developmental9"),("skills46","Communication","developmental9"),("skills47","Emergent Reading","developmental10"),("skills48","Emergent Writing","developmental10"),("skills49","Social-Emotional Learning","developmental11"),("skills50","Cognitive Development","developmental11"),("skills51","Physical Coordination","developmental11"),("skills52","Technology","developmental12"),("skills53","Engineering","developmental12"),("skills54","Miscellaneous","developmental12");
INSERT INTO `subjects` (`id`,`name`) VALUES ("subject1","Language/Literacy"),("subject2","Math"),("subject3","Science");
INSERT INTO `visibility_settings` (`id`,`name`) VALUES ("visibility_settings1","library_label_visibility_school"),("visibility_settings2","library_label_visibility_organization");
