package entity

type LockLog struct {
	ID         string `json:"id" dynamodbav:"id"`
	RecordID   string `json:"record_id" dynamodbav:"record_id"`
	OperatorID string `json:"operator_id" dynamodbav:"operator_id"`
	CreatedAt  string `json:"created_at" dynamodbav:"created_at"`
}
