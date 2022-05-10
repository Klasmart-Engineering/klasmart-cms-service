package main

import (
	"bufio"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"

	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
	"github.com/KL-Engineering/kidsloop-cms-service/da"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	"github.com/jinzhu/gorm"
)

func main() {
	cfg := &config.Config{
		DBConfig: config.DBConfig{
			//root:Passw0rd@tcp(127.0.0.1:3306)/kidsloop2?charset=utf8mb4&parseTime=True&loc=Local
			ConnectionString: "",
		},
	}
	ctx := context.Background()

	flag.StringVar(&cfg.DBConfig.ConnectionString, "dsn", "", `db connection string,required`)
	flag.Parse()

	if cfg.DBConfig.ConnectionString == "" {
		fmt.Println("Please enter params, --help")
		return
	}
	fmt.Println("dsn:", cfg.DBConfig.ConnectionString)
	fmt.Println("Enter to continue....")
	inputReader := bufio.NewReader(os.Stdin)
	inputReader.ReadString('\n')

	config.Set(cfg)
	option := dbo.WithConnectionString(cfg.DBConfig.ConnectionString)
	newDBO, err := dbo.NewWithConfig(option)
	if err != nil {
		log.Error(ctx, "dbo config error", log.Err(err))
		return
	}
	dbo.ReplaceGlobal(newDBO)
	orgIDs, err := GetScheduleAboutOrgIDs(context.Background())
	if err != nil {
		log.Error(ctx, "get org ids error", log.Err(err))
		return
	}
	if len(orgIDs) <= 0 {
		fmt.Println("not found org")
		return
	}
	fmt.Println("The org's in the schedules table are as follows")
	for len(orgIDs) > 0 {
		for _, id := range orgIDs {
			fmt.Println("orgID:", id)
		}

		var orgID = ""
		fmt.Println("Please enter the organisation id to be migrate or enter 'all' to migrate all organisation: ")
		fmt.Scanln(&orgID)
		if orgID == "all" {
			for _, item := range orgIDs {
				Migrate(item)
			}
			break
		}
		err = Migrate(orgID)
		if err != nil {
			continue
		}
		index := 0
		for i, id := range orgIDs {
			if id == orgID {
				index = i
			}
		}
		orgIDs = append(orgIDs[:index], orgIDs[(index+1):]...)
	}
	fmt.Println("Completed!")
}
func Migrate(orgID string) error {
	ctx := context.Background()
	fmt.Println(fmt.Sprintf("start migrate org(%s)", orgID))
	time.Sleep(2 * time.Second)
	total, err := StartMigrateByOrg(context.Background(), orgID)
	if err != nil {
		log.Error(ctx, "migrate org error", log.String("orgID", orgID), log.Err(err))
	}

	fmt.Println(fmt.Sprintf("migrate org(%s) success,count:%d", orgID, total))

	return err
}

func GetScheduleAboutOrgIDs(ctx context.Context) ([]string, error) {
	var data []*entity.Schedule
	err := dbo.MustGetDB(ctx).Table(constant.TableNameSchedule).
		Select("org_id").Group("org_id").Find(&data).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "GetScheduleAboutOrgList error", log.Err(err))
		return nil, err
	}
	orgIDs := make([]string, len(data))
	for i, item := range data {
		orgIDs[i] = item.OrgID
	}

	return orgIDs, nil
}

func StartMigrateByOrg(ctx context.Context, orgID string) (int64, error) {
	var scheduleList []*entity.Schedule
	err := da.GetScheduleDA().Query(ctx, da.ScheduleCondition{
		OrgID: sql.NullString{
			String: orgID,
			Valid:  true,
		},
	}, &scheduleList)
	if err != nil {
		log.Error(ctx, "query schedule by org error", log.Err(err), log.String("orgID", orgID))
		return 0, err
	}

	var scheduleRelations []*entity.ScheduleRelation

	for _, scheduleItem := range scheduleList {
		if scheduleItem.SubjectID != "" {
			scheduleRelations = append(scheduleRelations, &entity.ScheduleRelation{
				ID:           utils.NewID(),
				ScheduleID:   scheduleItem.ID,
				RelationID:   scheduleItem.SubjectID,
				RelationType: entity.ScheduleRelationTypeSubject,
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
			log.Error(ctx, "GetScheduleRelationDA error", log.Err(err), log.String("orgID", orgID))
			return 0, err
		}
		if count > 0 {
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
		log.Info(ctx, "For this organisation, there is no data to be migrated", log.Err(err), log.String("orgID", orgID))
		return 0, nil
	}
	rowCount, err := da.GetScheduleRelationDA().MultipleBatchInsert(ctx, dbo.MustGetDB(ctx), relationInsertData)
	return rowCount, err
}
