package service

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"strings"
)

type PropertySet struct {
	ID          string
	Program     string
	Subject     []string
	Category    string
	SubCategory []string
	Age         []string
	Grade       []string
}

func (p *PropertySet) ContentProperties() []*entity.ContentProperty {
	res := make([]*entity.ContentProperty, len(p.Subject)+len(p.SubCategory)+len(p.Age)+len(p.Grade)+2)
	index := 0

	res[index] = &entity.ContentProperty{
		PropertyType: entity.ContentPropertyTypeProgram,
		ContentID:    p.ID,
		PropertyID:   p.Program,
		Sequence:     0,
	}
	index++

	for i := range p.Subject {
		res[index] = &entity.ContentProperty{
			PropertyType: entity.ContentPropertyTypeSubject,
			ContentID:    p.ID,
			PropertyID:   p.Subject[i],
			Sequence:     i,
		}
		index++
	}

	res[index] = &entity.ContentProperty{
		PropertyType: entity.ContentPropertyTypeCategory,
		ContentID:    p.ID,
		PropertyID:   p.Category,
		Sequence:     0,
	}
	index++

	for i := range p.SubCategory {
		res[index] = &entity.ContentProperty{
			PropertyType: entity.ContentPropertyTypeSubCategory,
			ContentID:    p.ID,
			PropertyID:   p.SubCategory[i],
			Sequence:     i,
		}
		index++
	}

	for i := range p.Age {
		res[index] = &entity.ContentProperty{
			PropertyType: entity.ContentPropertyTypeAge,
			ContentID:    p.ID,
			PropertyID:   p.Age[i],
			Sequence:     i,
		}
		index++
	}

	for i := range p.Grade {
		res[index] = &entity.ContentProperty{
			PropertyType: entity.ContentPropertyTypeGrade,
			ContentID:    p.ID,
			PropertyID:   p.Grade[i],
			Sequence:     i,
		}
		index++
	}
	return res
}

type ContentObject struct {
	ID          string `gorm:"type:varchar(50);PRIMARY_KEY"`
	ContentType int    `gorm:"type:int;NOT NULL; column:content_type"`
	Name        string `gorm:"type:varchar(255);NOT NULL;column:content_name"`
	Keywords    string `gorm:"type:text;NOT NULL;column:keywords"`
	Description string `gorm:"type:text;NOT NULL;column:description"`
	Thumbnail   string `gorm:"type:text;NOT NULL;column:thumbnail"`

	SourceType string `gorm:"type:varchar(256); column:source_type"`

	Outcomes string `gorm:"type:text;NOT NULL;column:outcomes"`
	Data     string `gorm:"type:json;NOT NULL;column:data"`
	Extra    string `gorm:"type:text;NOT NULL;column:extra"`

	SuggestTime int    `gorm:"type:int;NOT NULL;column:suggest_time"`
	Author      string `gorm:"type:varchar(50);NOT NULL;column:author"`
	Creator     string `gorm:"type:varchar(50);NOT NULL;column:creator"`
	Org         string `gorm:"type:varchar(50);NOT NULL;column:org"`

	SelfStudy    int    `gorm:"type:tinyint;NOT NULL;column:self_study"`
	DrawActivity int    `gorm:"type:tinyint;NOT NULL;column:draw_activity"`
	LessonType   string `gorm:"type:varchar(100);column:lesson_type"`

	PublishStatus string `gorm:"type:varchar(16);NOT NULL;column:publish_status;index"`

	RejectReason string `gorm:"type:varchar(255);NOT NULL;column:reject_reason"`
	Remark       string `gorm:"type:varchar(255);NOT NULL;column:remark"`
	Version      int64  `gorm:"type:int;NOT NULL;column:version"`
	LockedBy     string `gorm:"type:varchar(50);NOT NULL;column:locked_by"`
	SourceID     string `gorm:"type:varchar(50);NOT NULL;column:source_id"`
	LatestID     string `gorm:"type:varchar(50);NOT NULL;column:latest_id"`

	CopySourceID string `gorm:"type:varchar(50);column:copy_source_id"`

	DirPath string `gorm:"type:varchar(2048);column:dir_path"`

	CreateAt int64 `gorm:"type:bigint;NOT NULL;column:create_at"`
	UpdateAt int64 `gorm:"type:bigint;NOT NULL;column:update_at"`
	DeleteAt int64 `gorm:"type:bigint;column:delete_at"`

	Program           string `gorm:"type:varchar(255);column:program" json:"program"`
	Subject           string `gorm:"type:varchar(255);column:subject" json:"subject"`
	Category          string `gorm:"type:varchar(255);column:developmental" json:"developmental"`
	SubCategory       string `gorm:"type:varchar(255);column:skills" json:"skills"`
	Age               string `gorm:"type:varchar(255);column:age" json:"age"`
	Grade             string `gorm:"type:varchar(255);column:grade" json:"grade"`
	VisibilitySetting string `gorm:"type:varchar(255);column:publish_scope" json:"publish_scope"`
}

type ContentObjectDA struct {
	s dbo.BaseDA
}

func (cd *ContentObjectDA) CreateContent(ctx context.Context, tx *dbo.DBContext, co ContentObject) (string, error) {
	co.ID = utils.NewID()
	_, err := cd.s.InsertTx(ctx, tx, &co)
	if err != nil {
		return "", err
	}
	return co.ID, nil
}

func (cd *ContentObjectDA) UpdateContent(ctx context.Context, tx *dbo.DBContext, cid string, co *ContentObject) error {
	co.ID = cid
	_, err := cd.s.UpdateTx(ctx, tx, co)
	if err != nil {
		return err
	}

	return nil
}
func (cd *ContentObjectDA) SearchContentInternal(ctx context.Context, tx *dbo.DBContext, condition *da.ContentConditionInternal) ([]*ContentObject, error) {
	objs := make([]*ContentObject, 0)
	err := cd.s.QueryTx(ctx, tx, condition, &objs)
	if err != nil {
		return nil, err
	}

	return objs, nil
}

type ContentService struct {
}

func (c *ContentService) Do(mapper Mapper) error {
	ctx := context.Background()
	contentList, propertiesMap, err := c.fetchContentData(ctx)
	if err != nil {
		return err
	}

	for i := range contentList {
		err := c.handleContent(ctx, mapper, contentList[i], propertiesMap[contentList[i].ID])
		if err != nil {
			fmt.Printf("Handle content %v failed\n", contentList[i].ID)
			continue
		}
	}

	return nil
}

func (c *ContentService) handleContent(ctx context.Context, mapper Mapper, content *ContentObject, properties []*entity.ContentProperty) error {
	var propertySet *PropertySet
	var err error
	if properties == nil {
		propertySet, err = c.doMappingOldContent(ctx, mapper, content)
	} else {
		propertySet, err = c.doMappingNewContent(ctx, mapper, content, properties)
	}
	if err != nil {
		return err
	}
	return c.insertNewPropertySet(ctx, propertySet)
}

func (c *ContentService) doMappingOldContent(ctx context.Context, mapper Mapper, content *ContentObject) (*PropertySet, error) {
	propertySet, err := c.doPropertyMapping(ctx, mapper, content.Org, &PropertySet{
		ID:          content.ID,
		Program:     content.Program,
		Subject:     strings.Split(content.Subject, ","),
		Category:    content.Category,
		SubCategory: strings.Split(content.SubCategory, ","),
		Age:         strings.Split(content.Age, ","),
		Grade:       strings.Split(content.Grade, ","),
	})
	if err != nil {
		return nil, err
	}

	return propertySet, nil
}
func (c *ContentService) doMappingNewContent(ctx context.Context, mapper Mapper, content *ContentObject, properties []*entity.ContentProperty) (*PropertySet, error) {
	propertySet := &PropertySet{
		ID: content.ID,
	}
	for i := range properties {
		switch properties[i].PropertyType {
		case entity.ContentPropertyTypeProgram:
			propertySet.Program = properties[i].PropertyID
		case entity.ContentPropertyTypeSubject:
			propertySet.Subject = append(propertySet.Subject, properties[i].PropertyID)
		case entity.ContentPropertyTypeCategory:
			propertySet.Category = properties[i].PropertyID
		case entity.ContentPropertyTypeAge:
			propertySet.Age = append(propertySet.Age, properties[i].PropertyID)
		case entity.ContentPropertyTypeGrade:
			propertySet.Grade = append(propertySet.Grade, properties[i].PropertyID)
		case entity.ContentPropertyTypeSubCategory:
			propertySet.SubCategory = append(propertySet.SubCategory, properties[i].PropertyID)
		}
	}
	return c.doPropertyMapping(ctx, mapper, content.Org, propertySet)
}

func (c *ContentService) insertNewPropertySet(ctx context.Context, propertySet *PropertySet) error {
	return dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		err := da.GetContentPropertyDA().CleanByContentID(ctx, tx, propertySet.ID)
		if err != nil {
			return err
		}
		err = da.GetContentPropertyDA().BatchAdd(ctx, tx, propertySet.ContentProperties())
		if err != nil {
			return err
		}
		return nil
	})
}

func (c *ContentService) doPropertyMapping(ctx context.Context, mapper Mapper, org string, propertySet *PropertySet) (*PropertySet, error) {
	newPropertySet := &PropertySet{
		ID: propertySet.ID,
	}

	//program
	newProgram, err := mapper.Program(ctx, org, propertySet.Program)
	if err != nil {
		return nil, err
	}
	newPropertySet.Program = newProgram

	//subjects
	newSubjects := make([]string, 0)
	for i := range propertySet.Subject {
		tempSubject, err := mapper.Subject(ctx, org, propertySet.Program, propertySet.Subject[i])
		if err != nil {
			return nil, err
		}
		newSubjects = append(newSubjects, tempSubject)
	}
	newPropertySet.Subject = newSubjects

	//category
	newCategory, err := mapper.Category(ctx, org, propertySet.Program, propertySet.Category)
	if err != nil {
		return nil, err
	}
	newPropertySet.Category = newCategory

	//sub category
	newSubCategories := make([]string, 0)
	for i := range propertySet.SubCategory {
		tempSubCategory, err := mapper.SubCategory(ctx, org, propertySet.Program, propertySet.Category, propertySet.SubCategory[i])
		if err != nil {
			return nil, err
		}
		newSubCategories = append(newSubCategories, tempSubCategory)
	}
	propertySet.SubCategory = newSubCategories

	//age
	newAges := make([]string, 0)
	for i := range propertySet.Age {
		tempAge, err := mapper.Age(ctx, org, propertySet.Program, propertySet.Age[i])
		if err != nil {
			return nil, err
		}
		newAges = append(newAges, tempAge)
	}
	propertySet.Age = newAges

	//grade
	newGrades := make([]string, 0)
	for i := range propertySet.Grade {
		tempGrade, err := mapper.Grade(ctx, org, propertySet.Program, propertySet.Grade[i])
		if err != nil {
			return nil, err
		}
		newGrades = append(newAges, tempGrade)
	}
	propertySet.Grade = newGrades

	return propertySet, nil
}

func (c *ContentService) fetchContentData(ctx context.Context) ([]*ContentObject, map[string][]*entity.ContentProperty, error) {
	contentDA := new(ContentObjectDA)
	contentList, err := contentDA.SearchContentInternal(ctx, dbo.MustGetDB(ctx), &da.ContentConditionInternal{})
	if err != nil {
		return nil, nil, err
	}
	ids := make([]string, len(contentList))
	for i := range contentList {
		ids[i] = contentList[i].ID
	}

	properties, err := da.GetContentPropertyDA().BatchGetByContentIDList(ctx, dbo.MustGetDB(ctx), ids)
	if err != nil {
		return nil, nil, err
	}
	propertiesMap := make(map[string][]*entity.ContentProperty)
	for i := range properties {
		propertiesMap[properties[i].ContentID] = append(propertiesMap[properties[i].ContentID], properties[i])
	}
	return contentList, propertiesMap, nil
}
