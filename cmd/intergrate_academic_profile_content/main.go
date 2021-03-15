package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/cmd/intergrate_academic_profile"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)


var (
	//connDBStr = "admin:LH1MCuL3V0Ib3254@tcp(kl2-migration-test.copqnkcbdsts.ap-northeast-2.rds.amazonaws.com:28344)/kidsloop2?parseTime=true&charset=utf8mb4"
	//connDBStr = "admin:LH1MCuL3V0Ib3254@tcp(192.168.1.233:28344)/kidsloop2?parseTime=true&charset=utf8mb4"
	connDBStr = "root:Badanamu123456@tcp(192.168.1.234:3306)/kidsloop2?parseTime=true&charset=utf8mb4"
	doUpdate  = true
	apiURL = "https://api.beta.kidsloop.net/user/"

	operator *entity.Operator = &entity.Operator{
		UserID: "14494c07-0d4f-5141-9db2-15799993f448", // PJ
		OrgID:  "10f38ce9-5152-4049-b4e7-6d2e2ba884e6", // Badanamu HQ
		Token:  "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE0NDk0YzA3LTBkNGYtNTE0MS05ZGIyLTE1Nzk5OTkzZjQ0OCIsImVtYWlsIjoicGoud2lsbGlhbXNAY2FsbWlkLmNvbSIsImV4cCI6MTYxNTc5MzAzMywiaXNzIjoia2lkc2xvb3AifQ.hXa9sQ3QzTMgBz_UtiRvoOoLfk_16reKp--iWSnZBO2qBX28fDvi_PvdNNqh2tq1AFv0XaUHzz59nQRWxbmXa_HvrwAhGjqBOdZL927yDiHKZI26enQlkkoNoJhak2n1IgUyG_q5kwvR0_5A_-DwBN6wGjQMqgvQjskI9zmFfkcsbE04X36Rhh3LQoSF7H7Trhy8R0aFO38gN7Q2Uuhnik2RJ-iEFxLdcv-5B9-QbKm4TSRgHwmenC5ocYWKohBLBM0heWFCfs2DnGERnr-uuj0_GWSBz5BtVVnF36xNGWWjuFttauEbDzCTuRtdoDg-Ft6mF43zgtOppFm8bpXwJg",
	}
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
		AMS: config.AMSConfig{
			EndPoint: apiURL,
		},
	})
	dbo.ReplaceGlobal(dboHandler)
	return nil
}

func loadContents(ctx context.Context) ([]*entity.Content, error) {
	_, contentList, err := da.GetContentDA().SearchContentInternal(ctx, dbo.MustGetDB(ctx), da.ContentConditionInternal{})
	if err != nil {
		return nil, err
	}
	return contentList, nil
}

func loadContentsPages(ctx context.Context) ([]*entity.Content, error) {
	res := make([]*entity.Content, 0)
	pageIndex := int64(1)
	pageSize := int64(1000)
	total, contentList, err := da.GetContentDA().SearchContent(ctx, dbo.MustGetDB(ctx), da.ContentCondition{
		Pager: utils.Pager{PageIndex: pageIndex, PageSize: pageSize},
	})
	if err != nil {
		return nil, err
	}
	res = append(res, contentList...)
	for len(res) < total {
		pageIndex ++
		total, contentList, err = da.GetContentDA().SearchContent(ctx, dbo.MustGetDB(ctx), da.ContentCondition{
			Pager: utils.Pager{PageIndex: pageIndex, PageSize: pageSize},
		})
		if err != nil {
			return nil, err
		}
		res = append(res, contentList...)
	}

	return contentList, nil
}


func mapContent(ctx context.Context, content *entity.Content, mapper intergrate_academic_profile.Mapper) error {
	//program map
	newProgram, err := mapper.Program(ctx, content.Org, content.Program)
	if err != nil {
		fmt.Println("can't find program:", content.Program)
		log.Error(ctx,
			"Can't find program",
			log.Err(err),
			log.Any("Program", content.Program),
			log.Any("Content", content))
		// return errors.New("Can't find program")
	}

	//subject map
	newSubjects := make([]string, 0)
	if content.Subject != "" {
		subjectsArray := strings.Split(content.Subject, ",")
		for i := range subjectsArray {
			newSubject, err := mapper.Subject(ctx, content.Org, content.Program, subjectsArray[i])
			if err != nil {
				fmt.Println("can't find subject:", subjectsArray[i], ", program:", content.Program)
				log.Error(ctx,
					"Can't find grade",
					log.Err(err),
					log.Any("Grade", content.Subject),
					log.Any("Content", content))
				// return errors.New("Can't find subject")
			 }
			newSubjects = append(newSubjects, newSubject)
		}
		newSubjects = utils.SliceDeduplication(newSubjects)
	}

	//developmental map
	newDevelopmental, err := mapper.Category(ctx, content.Org, content.Program, content.Developmental)
	if err != nil {
		fmt.Println("can't find developmental:", content.Developmental, ", program:", content.Program)
		log.Error(ctx,
			"Can't find developmental",
			log.Err(err),
			log.Any("Developmental", content.Developmental),
			log.Any("Content", content))
		// return errors.New("Can't find developmental")
	}

	//skills map
	newSkills := make([]string, 0)
	if content.Skills != "" {
		skillArray := strings.Split(content.Skills, ",")
		for i := range skillArray {
			newSkill, err := mapper.SubCategory(ctx, content.Org, content.Program, content.Developmental, skillArray[i])
			if err != nil {
				fmt.Println("can't find skill:", skillArray[i], ", program:", content.Program)
				log.Error(ctx,
					"Can't find skill",
					log.Err(err),
					log.Any("Skills", content.Skills),
					log.Any("Content", content))
				// return errors.New("Can't find skill")
			}
			newSkills = append(newSkills, newSkill)
		}
	}

	//ages map
	newAges := make([]string, 0)
	if content.Age != "" {
		ageArray := strings.Split(content.Age, ",")
		for i := range ageArray {
			newAge, err := mapper.Age(ctx, content.Org, content.Program, ageArray[i])
			if err != nil {
				fmt.Println("can't find age:", ageArray[i], ", program:", content.Program)
				log.Error(ctx,
					"Can't find age",
					log.Err(err),
					log.Any("Age", content.Age),
					log.Any("Content", content))
				// return errors.New("Can't find age")
			}
			newAges = append(newAges, newAge)
		}
	}

	//grades map
	newGrades := make([]string, 0)
	if content.Grade != "" {
		gradeArray := strings.Split(content.Grade, ",")
		for i := range gradeArray {
			newGrade, err := mapper.Grade(ctx, content.Org, content.Program, gradeArray[i])
			if err != nil {
				fmt.Println("can't find grade:", gradeArray[i], ", program:", content.Program)
				log.Error(ctx,
					"Can't find grade",
					log.Err(err),
					log.Any("Grade", content.Grade),
					log.Any("Content", content))
				// return errors.New("Can't find grade")
			}
			newGrades = append(newGrades, newGrade)
		}
	}

	content.Program = newProgram
	content.Subject = strings.Join(newSubjects, ",")
	content.Developmental = newDevelopmental
	content.Skills = strings.Join(newSkills, ",")
	content.Age = strings.Join(newAges, ",")
	content.Grade = strings.Join(newGrades, ",")

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
	errCount := int32(0)

	wg := new(sync.WaitGroup)
	poolChan := make(chan struct{}, 20)

	wg.Add(len(mappedIndex))
	for i := range mappedIndex {
		go func(i int) {
			poolChan <- struct{}{}
			defer func() {
				<-poolChan
				wg.Done()
			}()
			content := contentList[mappedIndex[i]]
			log.Info(ctx, "do update content", log.String("id", content.ID))
			err := da.GetContentDA().UpdateContent(ctx, dbo.MustGetDB(ctx), content.ID, *content)
			if err != nil {
				log.Warn(ctx,
					"content can't update",
					log.Err(err),
					log.Any("content", content),
					log.Int("index", mappedIndex[i]))
				atomic.AddInt32(&errCount, 1)
				return
			}
		}(i)
	}

	wg.Wait()
	if errCount > 0 {
		return errors.New("update parts of content failed")
	}
	return nil
}
func readParams(){
	args := os.Args
	if len(args) > 1 {
		connDBStr = args[1]
	}
	if len(args) > 2{
		apiURL = args[2]
	}

	if len(args) > 3{
		operator.Token = args[3]
	}
}

func requestToken() string{
	res, err := http.Get("http://192.168.1.233:10210/ll?email=pj.williams@calmid.com")
	if err != nil{
		panic(err)
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil{
		panic(err)
	}
	defer res.Body.Close()
	access := struct {
		Hit struct{
			Access string `json:"access"`
		} `json:"hit"`
	}{}
	err = json.Unmarshal(data, &access)
	if err != nil{
		panic(err)
	}
	return access.Hit.Access
}

func main() {
	readParams()
	ctx := context.Background()
	//打开数据库
	err := initDB(ctx, connDBStr)
	if err != nil {
		log.Error(ctx, "Can't open database", log.Err(err))
		return
	}
	//operator.Token = requestToken()

	//读取contents记录
	contentList, err := loadContents(ctx)
	if err != nil {
		log.Error(ctx, "Can't load contentList", log.Err(err))
		return
	}

	//mapper
	mappedIndex := mapper(ctx, contentList)
	log.Info(ctx, "contentList result:", log.Int("contentList size", len(contentList)))
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
