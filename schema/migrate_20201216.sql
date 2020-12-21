-- init headquarters: Badanamu HQ
insert into organizations_properties(id, type) values('0f38ce9-5152-4049-b4e7-6d2e2ba884e6', 'headquarters');

-- add column org_type for programs
ALTER TABLE `programs` ADD COLUMN org_type varchar(100) DEFAULT NULL COMMENT 'org_type';
ALTER TABLE `programs` ADD INDEX idx_org_type_delete_at ( `org_type`, `delete_at`);

-- update data for programs
update `programs` set org_type = 'headquarters' where id in ('program0','program1','program2','program3','program4','program5','program6','program7');
-- insert normal data
INSERT INTO `programs` (`id`, `name`, `create_id`, `update_id`, `delete_id`, `create_at`, `update_at`, `delete_at`, `number`, `org_type`) VALUES 
('5fd9ddface9660cbc5f667d8','None Specified','64a36ec1-7aa2-53ab-bb96-4c4ff752096b','','',1608113658,0,0,1000,'normal'),
('5fdac06ea878718a554ff00d','ESL','64a36ec1-7aa2-53ab-bb96-4c4ff752096b','','',1608171630,0,0,0,'normal'),
('5fdac0f61f066722a1351adb','Math','64a36ec1-7aa2-53ab-bb96-4c4ff752096b','','',1608171766,0,0,0,'normal'),
('5fdac0fe1f066722a1351ade','Science','64a36ec1-7aa2-53ab-bb96-4c4ff752096b','','',1608171774,0,0,0,'normal');

INSERT INTO `programs_ages` (`id`, `program_id`, `age_id`) VALUES 
('5fdace28842bcaf37f169ba8', '5fd9ddface9660cbc5f667d8', 'age0'),
('5fdac2eeae478ff90653f669', '5fdac06ea878718a554ff00d', 'age1'),
('5fdac2eeae478ff90653f66a', '5fdac06ea878718a554ff00d', 'age2'),
('5fdac2eeae478ff90653f66b', '5fdac06ea878718a554ff00d', 'age3'),
('5fdac2eeae478ff90653f66c', '5fdac06ea878718a554ff00d', 'age4'),
('5fdac2e2ae478ff90653f661', '5fdac0f61f066722a1351adb', 'age1'),
('5fdac2e2ae478ff90653f662', '5fdac0f61f066722a1351adb', 'age2'),
('5fdac2e2ae478ff90653f663', '5fdac0f61f066722a1351adb', 'age3'),
('5fdac2e2ae478ff90653f664', '5fdac0f61f066722a1351adb', 'age4'),
('5fdac2e2ae478ff90653f665', '5fdac0f61f066722a1351adb', 'age5'),
('5fdac4133f1e0c8bda5749ce', '5fdac0fe1f066722a1351ade', 'age1'),
('5fdac4133f1e0c8bda5749cf', '5fdac0fe1f066722a1351ade', 'age2'),
('5fdac4133f1e0c8bda5749d0', '5fdac0fe1f066722a1351ade', 'age3'),
('5fdac4133f1e0c8bda5749d1', '5fdac0fe1f066722a1351ade', 'age4'),
('5fdac4133f1e0c8bda5749d2', '5fdac0fe1f066722a1351ade', 'age5');

INSERT INTO `programs_developments` (`id`, `program_id`, `development_id`) VALUES 
('5fdace3f842bcaf37f169bb6', '5fd9ddface9660cbc5f667d8', 'developmental0'),
('5fdac25649757bb8ed19dbe0', '5fdac06ea878718a554ff00d', 'developmental1'),
('5fdac25649757bb8ed19dbe1', '5fdac06ea878718a554ff00d', 'developmental2'),
('5fdac25649757bb8ed19dbe2', '5fdac06ea878718a554ff00d', 'developmental3'),
('5fdac25649757bb8ed19dbe3', '5fdac06ea878718a554ff00d', 'developmental4'),
('5fdac25649757bb8ed19dbe4', '5fdac06ea878718a554ff00d', 'developmental5'),
('5fdac1831f066722a1351aff', '5fdac0f61f066722a1351adb', 'developmental1'),
('5fdac1831f066722a1351b00', '5fdac0f61f066722a1351adb', 'developmental2'),
('5fdac1831f066722a1351b01', '5fdac0f61f066722a1351adb', 'developmental3'),
('5fdac1831f066722a1351b02', '5fdac0f61f066722a1351adb', 'developmental4'),
('5fdac1831f066722a1351b03', '5fdac0f61f066722a1351adb', 'developmental5'),
('5fdac44d3f1e0c8bda5749e2', '5fdac0fe1f066722a1351ade', 'developmental1'),
('5fdac44d3f1e0c8bda5749e3', '5fdac0fe1f066722a1351ade', 'developmental2'),
('5fdac44d3f1e0c8bda5749e4', '5fdac0fe1f066722a1351ade', 'developmental3'),
('5fdac44d3f1e0c8bda5749e5', '5fdac0fe1f066722a1351ade', 'developmental4'),
('5fdac44d3f1e0c8bda5749e6', '5fdac0fe1f066722a1351ade', 'developmental5');

INSERT INTO `programs_grades` (`id`, `program_id`, `grade_id`) VALUES 
('5fdac1431f066722a1351aec', '5fdac06ea878718a554ff00d', 'grade2'),
('5fdace2f842bcaf37f169bae', '5fd9ddface9660cbc5f667d8', 'grade0'),
('5fdac1431f066722a1351aed', '5fdac06ea878718a554ff00d', 'grade3'),
('5fdac1431f066722a1351aee', '5fdac06ea878718a554ff00d', 'grade12'),
('5fdac1431f066722a1351aef', '5fdac06ea878718a554ff00d', 'grade5'),
('5fdac30bae478ff90653f672', '5fdac0f61f066722a1351adb', 'grade7'),
('5fdac30bae478ff90653f673', '5fdac0f61f066722a1351adb', 'grade8'),
('5fdac30bae478ff90653f674', '5fdac0f61f066722a1351adb', 'grade9'),
('5fdac30bae478ff90653f675', '5fdac0f61f066722a1351adb', 'grade10'),
('5fdac30bae478ff90653f676', '5fdac0f61f066722a1351adb', 'grade11'),
('5fdac4273f1e0c8bda5749d6', '5fdac0fe1f066722a1351ade', 'grade7'),
('5fdac4273f1e0c8bda5749d7', '5fdac0fe1f066722a1351ade', 'grade8'),
('5fdac4273f1e0c8bda5749d8', '5fdac0fe1f066722a1351ade', 'grade9'),
('5fdac4273f1e0c8bda5749d9', '5fdac0fe1f066722a1351ade', 'grade10'),
('5fdac4273f1e0c8bda5749da', '5fdac0fe1f066722a1351ade', 'grade11');

INSERT INTO `programs_subjects` (`id`, `program_id`, `subject_id`) VALUES 
('5fdace37842bcaf37f169bb2', '5fd9ddface9660cbc5f667d8', 'subject0'),
('5fdac1591f066722a1351af3', '5fdac06ea878718a554ff00d', 'subject1'),
('5fdac31aae478ff90653f67c', '5fdac0f61f066722a1351adb', 'subject2'),
('5fdac4303f1e0c8bda5749de', '5fdac0fe1f066722a1351ade', 'subject3');

INSERT INTO `developments_skills` (`id`, `development_id`, `skill_id`, `program_id`) VALUES 
('5fdace48842bcaf37f169bbb', 'developmental0', 'skills0', '5fd9ddface9660cbc5f667d8'),
('5fdac275ae478ff90653f646', 'developmental1', 'skills1', '5fdac06ea878718a554ff00d'),
('5fdac275ae478ff90653f647', 'developmental1', 'skills2', '5fdac06ea878718a554ff00d'),
('5fdac275ae478ff90653f648', 'developmental1', 'skills3', '5fdac06ea878718a554ff00d'),
('5fdac285ae478ff90653f64b', 'developmental2', 'skills7', '5fdac06ea878718a554ff00d'),
('5fdac291ae478ff90653f64e', 'developmental3', 'skills9', '5fdac06ea878718a554ff00d'),
('5fdac2a3ae478ff90653f651', 'developmental4', 'skills12', '5fdac06ea878718a554ff00d'),
('5fdac2bdae478ff90653f654', 'developmental5', 'skills14', '5fdac06ea878718a554ff00d'),
('5fdac2bdae478ff90653f655', 'developmental5', 'skills15', '5fdac06ea878718a554ff00d'),
('5fdac38349757bb8ed19dbf6', 'developmental1', 'skills20', '5fdac0f61f066722a1351adb'),
('5fdac38349757bb8ed19dbf0', 'developmental1', 'skills3', '5fdac0f61f066722a1351adb'),
('5fdac38349757bb8ed19dbf7', 'developmental1', 'skills37', '5fdac0f61f066722a1351adb'),
('5fdac38349757bb8ed19dbf5', 'developmental1', 'skills38', '5fdac0f61f066722a1351adb'),
('5fdac38349757bb8ed19dbf4', 'developmental1', 'skills39', '5fdac0f61f066722a1351adb'),
('5fdac38349757bb8ed19dbf3', 'developmental1', 'skills40', '5fdac0f61f066722a1351adb'),
('5fdac38349757bb8ed19dbf2', 'developmental1', 'skills41', '5fdac0f61f066722a1351adb'),
('5fdac38349757bb8ed19dbf1', 'developmental1', 'skills42', '5fdac0f61f066722a1351adb'),
('5fdac1c69c02dddf234c95b4', 'developmental2', 'skills7', '5fdac0f61f066722a1351adb'),
('5fdac3a63f1e0c8bda5749be', 'developmental3', 'skills10', '5fdac0f61f066722a1351adb'),
('5fdac3a63f1e0c8bda5749bf', 'developmental3', 'skills11', '5fdac0f61f066722a1351adb'),
('5fdac3a63f1e0c8bda5749bd', 'developmental3', 'skills9', '5fdac0f61f066722a1351adb'),
('5fdac3f13f1e0c8bda5749c3', 'developmental4', 'skills20', '5fdac0f61f066722a1351adb'),
('5fdac3f13f1e0c8bda5749c9', 'developmental4', 'skills3', '5fdac0f61f066722a1351adb'),
('5fdac3f13f1e0c8bda5749c2', 'developmental4', 'skills37', '5fdac0f61f066722a1351adb'),
('5fdac3f13f1e0c8bda5749c4', 'developmental4', 'skills38', '5fdac0f61f066722a1351adb'),
('5fdac3f13f1e0c8bda5749c5', 'developmental4', 'skills39', '5fdac0f61f066722a1351adb'),
('5fdac3f13f1e0c8bda5749c6', 'developmental4', 'skills40', '5fdac0f61f066722a1351adb'),
('5fdac3f13f1e0c8bda5749c7', 'developmental4', 'skills41', '5fdac0f61f066722a1351adb'),
('5fdac3f13f1e0c8bda5749c8', 'developmental4', 'skills42', '5fdac0f61f066722a1351adb'),
('5fdac1ff68a1a1ca0e48cb38', 'developmental5', 'skills14', '5fdac0f61f066722a1351adb'),
('5fdac1ff68a1a1ca0e48cb39', 'developmental5', 'skills15', '5fdac0f61f066722a1351adb'),
('5fdac4696b7a4c3c14177ff4', 'developmental1', 'skills40', '5fdac0fe1f066722a1351ade'),
('5fdac4696b7a4c3c14177ff5', 'developmental1', 'skills41', '5fdac0fe1f066722a1351ade'),
('5fdac4696b7a4c3c14177ff6', 'developmental1', 'skills42', '5fdac0fe1f066722a1351ade'),
('5fdac4786b7a4c3c14177ffb', 'developmental2', 'skills7', '5fdac0fe1f066722a1351ade'),
('5fdac4786b7a4c3c14177ffc', 'developmental2', 'skills8', '5fdac0fe1f066722a1351ade'),
('5fdac48e6b7a4c3c14178000', 'developmental3', 'skills10', '5fdac0fe1f066722a1351ade'),
('5fdac48f6b7a4c3c14178001', 'developmental3', 'skills11', '5fdac0fe1f066722a1351ade'),
('5fdac48e6b7a4c3c14177fff', 'developmental3', 'skills9', '5fdac0fe1f066722a1351ade'),
('5fdac4af6b7a4c3c14178009', 'developmental4', 'skills40', '5fdac0fe1f066722a1351ade'),
('5fdac4af6b7a4c3c1417800a', 'developmental4', 'skills41', '5fdac0fe1f066722a1351ade'),
('5fdac4af6b7a4c3c1417800b', 'developmental4', 'skills42', '5fdac0fe1f066722a1351ade'),
('5fdac4af6b7a4c3c1417800c', 'developmental4', 'skills43', '5fdac0fe1f066722a1351ade'),
('5fdac4c76b7a4c3c1417800f', 'developmental5', 'skills14', '5fdac0fe1f066722a1351ade'),
('5fdac4c76b7a4c3c14178010', 'developmental5', 'skills15', '5fdac0fe1f066722a1351ade');

