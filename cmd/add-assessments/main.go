package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	a, err := parseArgs()
	if err != nil {
		panic(err)
	}
	if err := a.check(); err != nil {
		panic(err)
	}
	if err := run(a); err != nil {
		panic(err)
	}
}

type args struct {
	baseURL     string
	orgID       string
	scheduleIDs []string
	cookie      string
}

func (a args) check() error {
	if a.baseURL == "" {
		return errors.New("require base url")
	}
	if a.orgID == "" {
		return errors.New("require org id")
	}
	if len(a.scheduleIDs) == 0 {
		return errors.New("require schedule ids")
	}
	if a.cookie == "" {
		return errors.New("require cookie")
	}
	return nil
}

func parseArgs() (args, error) {
	a := args{}
	flag.StringVar(&a.baseURL, "base-url", "https://kl2-test.kidsloop.net", "base url")
	flag.StringVar(&a.orgID, "org-id", "", "org id, required")
	flag.StringVar(&a.cookie, "cookie", "", "browser cookie, required")
	scheduleIDsString := flag.String("schedule-ids", "", "schedule ids, separate by comma, require one of \"schedule-ids\" and \"schedule-ids-file\"")
	scheduleIDsFile := flag.String("schedule-ids-file", "", "schedule ids file, separate by newline, require one of \"schedule-ids\" and \"schedule-ids-file\"")
	flag.Parse()
	if scheduleIDsString != nil && *scheduleIDsString != "" {
		items := strings.Split(*scheduleIDsString, ",")
		for _, item := range items {
			item := strings.TrimSpace(item)
			if item != "" {
				a.scheduleIDs = append(a.scheduleIDs, item)
			}
		}
	} else if scheduleIDsFile != nil && *scheduleIDsFile != "" {
		b, err := ioutil.ReadFile(*scheduleIDsFile)
		if err != nil {
			return args{}, err
		}
		items := strings.Split(string(b), "\n")
		for _, item := range items {
			item := strings.TrimSpace(item)
			if item != "" {
				a.scheduleIDs = append(a.scheduleIDs, item)
			}
		}
	}
	return a, nil
}

func run(a args) error {
	log.Printf("process args: %+v\n", a)
	for _, scheduleID := range a.scheduleIDs {
		log.Printf("processing: schedule_id = %s ...", scheduleID)
		if err := addAssessment(a.baseURL, a.orgID, scheduleID, a.cookie); err != nil {
			return err
		}
		log.Printf("process ok: schedule_id = %s", scheduleID)
	}
	return nil
}

func addAssessment(baseURL string, orgID string, scheduleID string, cookie string) error {
	schedule, err := getSchedule(baseURL, orgID, scheduleID, cookie)
	if err != nil {
		return err
	}
	studentIDs, err := getClassStudentIDs(schedule.Class.ID, cookie)
	if err != nil {
		return err
	}
	cmd := entity.AddAssessmentCommand{
		ScheduleID:    scheduleID,
		AttendanceIDs: studentIDs,
		ClassLength:   int(time.Unix(schedule.EndAt, 0).Sub(time.Unix(schedule.StartAt, 0)).Minutes()),
		ClassEndTime:  schedule.EndAt,
	}
	b, err := json.Marshal(cmd)
	if err != nil {
		panic(err)
		return err
	}
	req, err := http.NewRequest(http.MethodPost, baseURL+"/v1/assessments?org_id="+orgID, bytes.NewBuffer(b))
	if err != nil {
		panic(err)
		return err
	}
	req.Header.Add("Cookie", cookie)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
		return err
	}
	defer resp.Body.Close()
	b2, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
		return err
	}
	log.Println("process ok:", string(b2))
	return nil
}

func getSchedule(baseURL string, orgID string, scheduleID string, cookie string) (*entity.ScheduleDetailsView, error) {
	req, err := http.NewRequest(http.MethodGet, baseURL+"/v1/schedules/"+scheduleID+"?org_id="+orgID, nil)
	if err != nil {
		panic(err)
		return nil, err
	}
	req.Header.Add("Cookie", cookie)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
		return nil, err
	}
	defer resp.Body.Close()
	var result entity.ScheduleDetailsView
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		panic(err)
		return nil, err
	}
	return &result, nil
}

func getClassStudentIDs(classID string, cookie string) ([]string, error) {
	q := `query ($classID: ID!){
	class(class_id: $classID){
		students{
			id: user_id
			name: user_name
		}
  	}
}`
	req := chlorine.NewRequest(q)
	req.Header.Add("Cookie", cookie)
	req.Var("classID", classID)
	var students []*external.Student
	res := chlorine.Response{
		Data: &struct {
			Class struct {
				Students *[]*external.Student `json:"students"`
			} `json:"class"`
		}{Class: struct {
			Students *[]*external.Student `json:"students"`
		}{Students: &students}},
	}
	_, err := external.GetChlorine().Run(context.Background(), req, &res)
	if err != nil {
		return nil, err
	}
	if len(res.Errors) > 0 {
		return nil, res.Errors
	}

	result := make([]string, 0, len(students))
	for _, student := range students {
		result = append(result, student.ID)
	}

	return result, nil
}
