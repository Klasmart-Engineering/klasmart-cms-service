package main

import (
	"bufio"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/config"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
)

func main() {
	cfg := &config.Config{
		DBConfig: config.DBConfig{
			//root:Passw0rd@tcp(127.0.0.1:3306)/kidsloop2?charset=utf8mb4&parseTime=True&loc=Local
			ConnectionString: "",
		},
		AMS: config.AMSConfig{
			//https://api.kidsloop.net/user/
			EndPoint: "",
		},
	}
	operator := &entity.Operator{
		Token: "test",
	}

	flag.StringVar(&cfg.DBConfig.ConnectionString, "dsn", "", `db connection string,required`)
	flag.StringVar(&cfg.AMS.EndPoint, "ams", "", "AMS EndPoint,required")
	flag.Parse()

	if cfg.DBConfig.ConnectionString == "" || cfg.AMS.EndPoint == "" {
		fmt.Println("Please enter params, --help")
		return
	}
	fmt.Println("dsn:", cfg.DBConfig.ConnectionString)
	fmt.Println("ams:", cfg.AMS.EndPoint)
	fmt.Println("Enter to continue....")
	inputReader := bufio.NewReader(os.Stdin)
	inputReader.ReadString('\n')

	config.Set(cfg)
	option := dbo.WithConnectionString(cfg.DBConfig.ConnectionString)
	newDBO, err := dbo.NewWithConfig(option)
	if err != nil {
		log.Println("connection mysql db error:", err)
		return
	}
	dbo.ReplaceGlobal(newDBO)
	orgList, err := GetScheduleAboutOrgList(context.Background(), operator)
	if err != nil {
		log.Println("get schedule group by org error:", err)
		return
	}
	if len(orgList) <= 0 {
		fmt.Println("not found org")
		return
	}
	fmt.Println("The org's in the schedules table are as follows")
	for len(orgList) > 0 {
		for _, org := range orgList {
			fmt.Println("orgID:", org.ID, "orgName", org.Name)
		}

		var orgID = ""
		fmt.Println("Please enter the organisation id to be migrate or enter 'all' to migrate all organisation: ")
		fmt.Scanln(&orgID)
		if orgID == "all" {
			for _, item := range orgList {
				Migrate(item.ID, operator)
			}
			break
		}
		err = Migrate(orgID, operator)
		if err != nil {
			continue
		}
		index := 0
		for i, org := range orgList {
			if org.ID == orgID {
				index = i
			}
		}
		orgList = append(orgList[:index], orgList[(index+1):]...)
	}
	fmt.Println("Completed!")
}
func Migrate(orgID string, operator *entity.Operator) error {
	fmt.Println(fmt.Sprintf("start migrate org(%s)", orgID))
	time.Sleep(2 * time.Second)
	total, err := StartMigrateByOrg(context.Background(), operator, orgID)
	if err != nil {
		log.Println("Start Migrate By Org error,", err)
	}

	fmt.Println(fmt.Sprintf("migrate org(%s) success,count:%d", orgID, total))
	//fmt.Println("Enter to continue")
	//inputReader := bufio.NewReader(os.Stdin)
	//inputReader.ReadString('\n')
	return err
}

func GetScheduleAboutOrgList(ctx context.Context, op *entity.Operator) ([]*external.NullableOrganization, error) {
	var data []*entity.Schedule
	err := dbo.MustGetDB(ctx).Table(constant.TableNameSchedule).
		Select("org_id").Group("org_id").Find(&data).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Println("GetScheduleAboutOrgList error ", err)
		return nil, err
	}
	orgIDs := make([]string, len(data))
	for i, item := range data {
		orgIDs[i] = item.OrgID
	}

	orgInfo, err := external.GetOrganizationServiceProvider().BatchGet(ctx, op, orgIDs)
	if err != nil {
		log.Println("get org error ", err)
		return nil, err
	}

	return orgInfo, nil
}

func StartMigrateByOrg(ctx context.Context, op *entity.Operator, orgID string) (int64, error) {
	var scheduleList []*entity.Schedule
	err := da.GetScheduleDA().Query(ctx, da.ScheduleCondition{
		OrgID: sql.NullString{
			String: orgID,
			Valid:  true,
		},
	}, &scheduleList)
	if err != nil {
		log.Printf("query schedule by org error:%v,orgID:%s \n", err, orgID)
		return 0, err
	}

	var scheduleRelations []*entity.ScheduleRelation
	classIDs := make([]string, 0)
	for _, scheduleItem := range scheduleList {
		if scheduleItem.ClassID == "" {
			log.Printf("schedule class id is empty,scheduleID:%s \n", scheduleItem.ID)
			continue
		}
		classIDs = append(classIDs, scheduleItem.ClassID)
	}
	classIDs = utils.SliceDeduplication(classIDs)
	classOrgMap, err := external.GetOrganizationServiceProvider().GetByClasses(ctx, op, classIDs)
	if err != nil {
		log.Println("GetOrganizationServiceProvider.GetByClasses,classIDs:", classIDs)
		return 0, err
	}
	classSchoolMap, err := external.GetSchoolServiceProvider().GetByClasses(ctx, op, classIDs)
	if err != nil {
		log.Println("GetSchoolServiceProvider.GetByClasses,classIDs:", classIDs)
		return 0, err
	}
	classTeacherMap, err := external.GetTeacherServiceProvider().GetByClasses(ctx, op, classIDs)
	if err != nil {
		log.Println("GetTeacherServiceProvider.GetByClasses,classIDs:", classIDs)
		return 0, err
	}
	classStudentMap, err := external.GetStudentServiceProvider().GetByClassIDs(ctx, op, classIDs)
	if err != nil {
		log.Println("GetStudentServiceProvider.GetByClasses,classIDs:", classIDs)
		return 0, err
	}
	for _, scheduleItem := range scheduleList {
		_, ok := classOrgMap[scheduleItem.ClassID]
		schoolList, ok2 := classSchoolMap[scheduleItem.ClassID]
		// Validate classID
		if !ok && !ok2 {
			log.Printf("class id is invalid,classID:%v \n", scheduleItem.ClassID)
			continue
		}

		// org relation
		scheduleRelations = append(scheduleRelations, &entity.ScheduleRelation{
			ID:           utils.NewID(),
			ScheduleID:   scheduleItem.ID,
			RelationID:   orgID,
			RelationType: entity.ScheduleRelationTypeOrg,
		})
		// school relation
		for _, item := range schoolList {
			scheduleRelations = append(scheduleRelations, &entity.ScheduleRelation{
				ID:           utils.NewID(),
				ScheduleID:   scheduleItem.ID,
				RelationID:   item.ID,
				RelationType: entity.ScheduleRelationTypeSchool,
			})
		}
		// class relation
		scheduleRelations = append(scheduleRelations, &entity.ScheduleRelation{
			ID:           utils.NewID(),
			ScheduleID:   scheduleItem.ID,
			RelationID:   scheduleItem.ClassID,
			RelationType: entity.ScheduleRelationTypeClassRosterClass,
		})
		// teacher relation
		teacherList, _ := classTeacherMap[scheduleItem.ClassID]
		for _, teacher := range teacherList {
			scheduleRelations = append(scheduleRelations, &entity.ScheduleRelation{
				ID:           utils.NewID(),
				ScheduleID:   scheduleItem.ID,
				RelationID:   teacher.ID,
				RelationType: entity.ScheduleRelationTypeClassRosterTeacher,
			})
		}

		// student relation
		studentList, _ := classStudentMap[scheduleItem.ClassID]
		for _, stu := range studentList {
			scheduleRelations = append(scheduleRelations, &entity.ScheduleRelation{
				ID:           utils.NewID(),
				ScheduleID:   scheduleItem.ID,
				RelationID:   stu.ID,
				RelationType: entity.ScheduleRelationTypeClassRosterStudent,
			})
		}

	}
	var relationInsertData = make([]*entity.ScheduleRelation, 0)
	for _, item := range scheduleRelations {
		condition := &da.ScheduleRelationCondition{
			ScheduleID: sql.NullString{
				String: item.ScheduleID,
				Valid:  true,
			},
			RelationID: sql.NullString{
				String: item.RelationID,
				Valid:  true,
			},
		}
		count, err := da.GetScheduleRelationDA().Count(ctx, condition, &entity.ScheduleRelation{})
		if err != nil {
			log.Println("GetScheduleRelationDA.Count error,condition")
			return 0, err
		}
		if count > 0 {
			log.Println("schedule relation record already exist,condition")
			continue
		}
		relationInsertData = append(relationInsertData, &entity.ScheduleRelation{
			ID:           item.ID,
			ScheduleID:   item.ScheduleID,
			RelationID:   item.RelationID,
			RelationType: item.RelationType,
		})
	}

	if len(relationInsertData) <= 0 {
		log.Println("For this organisation, there is no data to be migrated")
		return 0, nil
	}

	rowCount, err := da.GetScheduleRelationDA().MultipleBatchInsert(ctx, dbo.MustGetDB(ctx), relationInsertData)
	return rowCount, err
}
