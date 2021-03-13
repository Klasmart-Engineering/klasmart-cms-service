package main

import (
	"context"
	"errors"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/cmd/intergrate_academic_profile"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
)

const (
	connDBStr = "root:Badanamu123456@tcp(192.168.1.234:3310)/kidsloop2?charset=utf8mb4&parseTime=True&loc=Local"
	doUpdate  = false
)

var (
	operator *entity.Operator = &entity.Operator{}
)

func initDB(ctx context.Context, str string) error {
	dboHandler, err := dbo.NewWithConfig(func(c *dbo.Config) {
		c.ShowLog = true
		c.ShowSQL = true
		c.MaxIdleConns = 2
		c.MaxOpenConns = 4
		c.ConnectionString = str
	})
	if err != nil {
		log.Error(ctx, "create dbo failed", log.Err(err))
		return err
	}
	config.Set(&config.Config{
		RedisConfig: config.RedisConfig{
			OpenCache: false,
			Host:      "",
			Port:      0,
			Password:  "",
		},
	})
	dbo.ReplaceGlobal(dboHandler)
	return nil
}

func loadContents(ctx context.Context) ([]*entity.Content, error) {
	_, contentList, err := da.GetContentDA().SearchContent(ctx, dbo.MustGetDB(ctx), da.ContentCondition{})
	if err != nil {
		return nil, err
	}
	return contentList, nil
}

func mapContent(ctx context.Context, content *entity.Content, mapper intergrate_academic_profile.Mapper) error {
	//program map
	newProgram, err := mapper.Program(ctx, content.Org, content.Program)
	if err != nil {
		log.Error(ctx,
			"Can't find program",
			log.Err(err),
			log.Any("Program", content.Program),
			log.Any("Content", content))
		return errors.New("Can't find program")
	}
	content.Program = newProgram

	//subject map
	newSubject, err := mapper.Subject(ctx, content.Org, content.Program, content.Subject)
	if err != nil {
		log.Error(ctx,
			"Can't find subject",
			log.Err(err),
			log.Any("Subject", content.Subject),
			log.Any("Content", content))
		return errors.New("Can't find subject")
	}
	content.Subject = newSubject

	//developmental map
	newDevelopmental, err := mapper.Category(ctx, content.Org, content.Program, content.Developmental)
	if err != nil {
		log.Error(ctx,
			"Can't find developmental",
			log.Err(err),
			log.Any("Developmental", content.Developmental),
			log.Any("Content", content))
		return errors.New("Can't find developmental")
	}
	content.Developmental = newDevelopmental

	//skills map
	if content.Skills != "" {
		newSkills := make([]string, 0)
		skillArray := strings.Split(content.Skills, ",")
		for i := range skillArray {
			newSkill, err := mapper.SubCategory(ctx, content.Org, content.Program, content.Developmental, skillArray[i])
			if err != nil {
				log.Error(ctx,
					"Can't find skill",
					log.Err(err),
					log.Any("Skills", content.Skills),
					log.Any("Content", content))
				return errors.New("Can't find skill")
			}
			newSkills = append(newSkills, newSkill)
		}
		content.Skills = strings.Join(newSkills, ",")
	}

	//ages map
	if content.Age != "" {
		newAges := make([]string, 0)
		ageArray := strings.Split(content.Age, ",")
		for i := range ageArray {
			newAge, err := mapper.Age(ctx, content.Org, content.Program, ageArray[i])
			if err != nil {
				log.Error(ctx,
					"Can't find age",
					log.Err(err),
					log.Any("Age", content.Age),
					log.Any("Content", content))
				return errors.New("Can't find age")
			}
			newAges = append(newAges, newAge)
		}
		content.Age = strings.Join(newAges, ",")
	}

	//grades map
	if content.Grade != "" {
		newGrades := make([]string, 0)
		gradeArray := strings.Split(content.Grade, ",")
		for i := range gradeArray {
			newGrade, err := mapper.Grade(ctx, content.Org, content.Program, gradeArray[i])
			if err != nil {
				log.Error(ctx,
					"Can't find grade",
					log.Err(err),
					log.Any("Grade", content.Grade),
					log.Any("Content", content))
				return errors.New("Can't find grade")
			}
			newGrades = append(newGrades, newGrade)
		}
		content.Grade = strings.Join(newGrades, ",")
	}

	return nil
}

func mapper(ctx context.Context, contentList []*entity.Content) []int {
	mapper := intergrate_academic_profile.NewMapper(operator)

	mappedIndex := make([]int, 0)
	for i := range contentList {
		err := mapContent(ctx, contentList[i], mapper)
		if err != nil {
			log.Warn(ctx,
				"content can't map",
				log.Err(err),
				log.Any("content", contentList[i]),
				log.Int("index", i))
			continue
		}
		mappedIndex = append(mappedIndex, i)
	}
	return mappedIndex
}

func updateContent(ctx context.Context, contentList []*entity.Content, mappedIndex []int) error {
	log.Info(ctx, "update content", log.Int("size", len(mappedIndex)))
	errCount := 0
	for i := range mappedIndex {
		content := contentList[mappedIndex[i]]
		err := da.GetContentDA().UpdateContent(ctx, dbo.MustGetDB(ctx), content.ID, *content)
		if err != nil {
			log.Warn(ctx,
				"content can't update",
				log.Err(err),
				log.Any("content", content),
				log.Int("index", mappedIndex[i]))
			errCount++
			continue
		}
	}
	if errCount > 0 {
		return errors.New("update parts of content failed")
	}
	return nil
}

func main() {
	ctx := context.Background()
	//打开数据库
	err := initDB(ctx, connDBStr)
	if err != nil {
		log.Error(ctx, "Can't open database", log.Err(err))
		return
	}
	//读取contents记录
	contentList, err := loadContents(ctx)
	if err != nil {
		log.Error(ctx, "Can't load contentList", log.Err(err))
		return
	}
	//mapper
	mappedIndex := mapper(ctx, contentList)
	log.Info(ctx, "mapper result:", log.Int("mapper size", len(mappedIndex)))

	if doUpdate {
		//更新contents记录
		err = updateContent(ctx, contentList, mappedIndex)
		if err != nil {
			log.Error(ctx, "Update content failed", log.Err(err))
			return
		}
		log.Info(ctx, "Done.")
	}

}
