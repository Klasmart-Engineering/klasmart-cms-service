create table `assessments`
(
    `id`             varchar(64)   not null comment 'id',
    `schedule_id`    varchar(64)   not null comment 'schedule id',
    `title`          varchar(1024) not null comment 'title',
    `program_id`     varchar(64)   not null comment 'program id',
    `subject_id`     varchar(64)   not null comment 'subject id',
    `teacher_id`     varchar(64)   not null comment 'teacher id',
    `class_length`   int           not null comment 'class length (util: minute)',
    `class_end_time` bigint        not null comment 'class end time (unix seconds)',
    `complete_time`  bigint        not null comment 'complete time (unix seconds)',
    `status`         varchar(128)  not null comment 'status (enum: in_progress, complete)',
    `create_time`    bigint        not null comment 'create time (unix seconds)',
    `update_time`    bigint        not null comment 'update time (unix seconds)',
    `delete_time`    bigint        not null comment 'delete time (unix seconds)',
    primary key (`id`),
    key `assessments_status` (status),
    key `assessments_schedule_id` (schedule_id),
    key `assessments_complete_time` (complete_time)
) comment 'assessments';

create table assessments_attendances
(
    `id`            varchar(64) not null comment 'id',
    `assessment_id` varchar(64) not null comment 'assessment id',
    `attendance_id` varchar(64) not null comment 'attendance id',
    primary key (`id`),
    key `assessments_attendances_assessment_id` (`assessment_id`),
    key `assessments_attendances_attendance_id` (`attendance_id`)
);

create table assessments_outcomes
(
    `id`            varchar(64) not null comment 'id',
    `assessment_id` varchar(64) not null comment 'assessment id',
    `outcome_id` varchar(64) not null comment 'outcome id',
    primary key (`id`),
    key `assessments_outcomes_assessment_id` (`assessment_id`),
    key `assessments_outcomes_outcome_id` (`outcome_id`)
) comment 'assessment and outcome map';

create table outcomes_attendances
(
    `id`            varchar(64) not null comment 'id',
    `assessment_id` varchar(64) not null comment 'assessment id',
    `outcome_id` varchar(64) not null comment 'outcome id',
    `attendance_id` varchar(64) not null comment 'attendance id',
    primary key (`id`),
    key `outcomes_attendances_assessment_id` (`outcome_id`),
    key `outcomes_attendances_outcome_id` (`outcome_id`),
    key `outcomes_attendances_attendance_id` (`attendance_id`)
) comment 'outcome and attendance map';
