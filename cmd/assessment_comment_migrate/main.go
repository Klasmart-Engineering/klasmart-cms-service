package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/dbo"
	"github.com/KL-Engineering/kidsloop-cms-service/config"
	"github.com/KL-Engineering/kidsloop-cms-service/da/assessmentV2"
	"github.com/KL-Engineering/kidsloop-cms-service/entity"
	v2 "github.com/KL-Engineering/kidsloop-cms-service/entity/v2"
	"github.com/KL-Engineering/kidsloop-cms-service/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strings"

	"time"
)

//CREATE TABLE assessment_xapi_teacher_comment(
//room_id 	varchar     NOT NULL,
//teacher_id  varchar     NOT NULL,
//student_id  varchar     NOT NULL,
//created_at  timestamptz,
//updated_at  timestamptz,
//comment		varchar,
//version		integer
//);

type TeacherComment struct {
	RoomID    string    `gorm:"room_id"`
	TeacherID string    `gorm:"teacher_id"`
	StudentID string    `gorm:"student_id"`
	CreatedAt time.Time `gorm:"created_at"`
	UpdatedAt time.Time `gorm:"updated_at"`
	Comment   string    `gorm:"comment"`
	Version   int64     `gorm:"version"`
}

func (TeacherComment) TableName() string {
	return "assessment_xapi_teacher_comment"
}

var (
	postgresDB *gorm.DB
)

func main() {
	ctx := context.Background()
	cms_mysql_dsn := ""
	assessment_postgres_dsn := ""
	// parse var
	flag.StringVar(&cms_mysql_dsn, "cms_dsn", "", `Mysql database connection string for cms-service,required`)
	flag.StringVar(&assessment_postgres_dsn, "assessment_dsn", "", `postgres database connection string for assessment-service,required`)
	flag.Parse()

	if cms_mysql_dsn == "" || cms_mysql_dsn == "" {
		fmt.Println("Please enter params, --help")
		return
	}

	// connect mysql
	err := initMysqlConfig(&args{
		DSN: cms_mysql_dsn, //"root:Passw0rd@tcp(127.0.0.1:3306)/kidsloop_alpha_cms?charset=utf8mb4&parseTime=True&loc=Local",
	})
	if err != nil {
		fmt.Println("init cms db connection error,", err)
		return
	}

	// connect postgres
	postgresDB, err = initPostgresConfig(&args{
		DSN: assessment_postgres_dsn, // "host=localhost user=postgres password=123456 dbname=assessmentdb port=5432 sslmode=disable TimeZone=Asia/Shanghai",
	})
	if err != nil {
		log.Panic(ctx, "connect postgres error", log.Err(err))
		return
	}

	var comments []*TeacherComment
	err = postgresDB.Find(&comments).Error

	// get assessment
	scheduleIDs := make([]string, 0)
	dupMap := make(map[string]struct{})
	for _, item := range comments {
		if _, ok := dupMap[item.RoomID]; ok {
			continue
		}
		dupMap[item.RoomID] = struct{}{}
		scheduleIDs = append(scheduleIDs, item.RoomID)
	}

	var assessments []*v2.Assessment
	condition := &assessmentV2.AssessmentCondition{
		ScheduleIDs: entity.NullStrings{
			Strings: scheduleIDs,
			Valid:   true,
		},
	}
	err = assessmentV2.GetAssessmentDA().Query(ctx, condition, &assessments)
	if err != nil {
		fmt.Println(err)
		return
	}

	// get assessment user
	assessmentIDs := make([]string, len(assessments))
	assessmentMap := make(map[string]*v2.Assessment)
	for i, item := range assessments {
		assessmentIDs[i] = item.ID
		assessmentMap[item.ID] = item
	}

	var assessmentUsers []*v2.AssessmentUser
	auCondition := &assessmentV2.AssessmentUserCondition{
		AssessmentIDs: entity.NullStrings{
			Strings: assessmentIDs,
			Valid:   false,
		},
		UserType: sql.NullString{
			String: v2.AssessmentUserTypeStudent.String(),
			Valid:  true,
		},
	}
	err = assessmentV2.GetAssessmentUserDA().Query(ctx, auCondition, &assessmentUsers)
	if err != nil {
		fmt.Println(err)
		return
	}
	assessmentUserMap := make(map[string]string, len(assessmentUsers))
	assessmentUserIDs := make([]string, 0, len(assessmentUsers))
	for _, item := range assessmentUsers {
		assessment, ok := assessmentMap[item.AssessmentID]
		if !ok {
			continue
		}
		assessmentUserIDs = append(assessmentUserIDs, item.ID)
		assessmentUserMap[GetAssessmentKey([]string{assessment.ScheduleID, item.UserID})] = item.ID
	}

	// get reviewer feedback
	var existFeedbacks []*v2.AssessmentReviewerFeedback
	feedbackCondition := assessmentV2.AssessmentUserResultCondition{AssessmentUserIDs: entity.NullStrings{
		Strings: assessmentUserIDs,
		Valid:   true,
	}}
	err = assessmentV2.GetAssessmentUserResultDA().Query(ctx, feedbackCondition, &existFeedbacks)
	if err != nil {
		log.Error(ctx, "query feedback error", log.Err(err))
		return
	}
	existFeedbackMap := make(map[string]struct{}, len(existFeedbacks))
	for _, item := range existFeedbacks {
		existFeedbackMap[item.AssessmentUserID] = struct{}{}
	}

	feedbacks := make([]*v2.AssessmentReviewerFeedback, 0)
	for _, item := range comments {
		assessmentUserID, ok := assessmentUserMap[GetAssessmentKey([]string{item.RoomID, item.StudentID})]
		if !ok {
			continue
		}
		if _, ok := existFeedbackMap[assessmentUserID]; ok {
			continue
		}

		feedItem := &v2.AssessmentReviewerFeedback{
			ID:                utils.NewID(),
			AssessmentUserID:  assessmentUserID,
			ReviewerID:        item.TeacherID,
			StudentFeedbackID: "",
			AssessScore:       0,
			ReviewerComment:   item.Comment,
			CreateAt:          item.CreatedAt.Unix(),
			UpdateAt:          item.UpdatedAt.Unix(),
			DeleteAt:          0,
		}
		feedbacks = append(feedbacks, feedItem)
	}

	_, err = assessmentV2.GetAssessmentUserResultDA().InsertInBatches(ctx, feedbacks, 400)
	if err != nil {
		log.Error(ctx, "insert reviewer feedback error", log.Err(err))
		return
	}

	log.Info(ctx, "migrate teacher comments success!", log.Err(err))
}

//func initComments(ctx context.Context) {
//	var assessments []*v2.Assessment
//	condition := &assessmentV2.AssessmentCondition{
//
//		Pager: utils.GetDboPagerFromInt(1, 1000),
//	}
//	err := assessmentV2.GetAssessmentDA().Query(ctx, condition, &assessments)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//
//	assessmentIDs := make([]string, len(assessments))
//	assessmentMap := make(map[string]*v2.Assessment)
//	for i, item := range assessments {
//		assessmentIDs[i] = item.ID
//		assessmentMap[item.ID] = item
//	}
//
//	var assessmentUsers []*v2.AssessmentUser
//	assessmentV2.GetAssessmentUserDA().Query(ctx, &assessmentV2.AssessmentUserCondition{
//		AssessmentIDs: entity.NullStrings{
//			Strings: assessmentIDs,
//			Valid:   true,
//		},
//	}, &assessmentUsers)
//
//	var comments = make([]*TeacherComment, 0)
//	for _, item := range assessmentUsers {
//		if item.UserType != v2.AssessmentUserTypeStudent {
//			continue
//		}
//		assessment, ok := assessmentMap[item.AssessmentID]
//		if !ok {
//			continue
//		}
//		commentItem := &TeacherComment{
//			RoomID:    assessment.ScheduleID,
//			TeacherID: item.UserID,
//			StudentID: item.UserID,
//			CreatedAt: time.Now(),
//			UpdatedAt: time.Now(),
//			Comment:   fmt.Sprintf("%s+%s", assessment.Title, assessment.AssessmentType),
//			Version:   1,
//		}
//		comments = append(comments, commentItem)
//	}
//	err = postgresDB.Model(&TeacherComment{}).CreateInBatches(comments, 200).Error
//	if err != nil {
//		log.Error(ctx, "inert error", log.Err(err))
//		return
//	}
//}

type args struct {
	DSN string `json:"dsn"`
}

func initPostgresConfig(a *args) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(a.DSN), &gorm.Config{})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return db, nil
}

func initMysqlConfig(a *args) error {
	// set config
	c := &config.Config{
		DBConfig: config.DBConfig{ConnectionString: a.DSN},
	}
	config.Set(c)

	// replace dbo
	newDBO, err := dbo.NewWithConfig(dbo.WithConnectionString(c.DBConfig.ConnectionString))
	if err != nil {
		return err
	}
	dbo.ReplaceGlobal(newDBO)

	fmt.Println("=> Init config done!")

	return nil
}

func GetAssessmentKey(value []string) string {
	return strings.Join(value, "_")
}
