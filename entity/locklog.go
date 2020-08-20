package entity

type LockLog struct {
	ID         string `json:"id" dynamodbav:"id"`
	RecordID   string `json:"record_id" dynamodbav:"record_id"`
	OperatorID string `json:"operator_id" dynamodbav:"operator_id"`
	CreatedAt  int64  `json:"created_at" dynamodbav:"created_at"`
	DeletedAt  int64  `json:"deleted_at" dynamodbav:"deleted_at"`
}

func (LockLog) TableName() string {
	return "lock_logs"
}

func (LockLog) PrimaryKey() []string {
	return []string{"id"}
}

func (LockLog) GlobalSecondIndexes() [][]string {
	return [][]string{
		{"record_id", "created_at"},
	}
}

func (LockLog) IndexNameOfRecordIDAndCreatedAt() string {
	return "record_id_and_created_at"
}
