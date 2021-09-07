package entity

type Modeler interface {
	TableName() string
}

type BatchInsertModeler interface {
	Modeler
	GetBatchInsertColsAndValues() (cols []string, values []interface{})
}
