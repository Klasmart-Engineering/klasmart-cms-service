
CREATE INDEX idx_org_id_delete_at_program_id ON schedules ( org_id, delete_at, program_id );

DROP INDEX idx_schedule_id_relation_type ON schedules_relations;

CREATE INDEX idx_relation_type ON schedules_relations ( relation_type  );

