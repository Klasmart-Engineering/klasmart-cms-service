package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/chlorine"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	log.Println("all done!")
}

type args struct {
	BaseURL         string   `json:"base_url"`
	GraphqlEndpoint string   `json:"graphql_endpoint"`
	OrgID           string   `json:"org_id"`
	ScheduleIDs     []string `json:"schedule_ids"`
	Cookie          string   `json:"cookie"`
}

func (a args) check() error {
	if a.BaseURL == "" {
		return errors.New("require base url")
	}
	if a.GraphqlEndpoint == "" {
		return errors.New("require graphql endpoint")
	}
	if a.OrgID == "" {
		return errors.New("require org id")
	}
	if len(a.ScheduleIDs) == 0 {
		return errors.New("require schedule ids")
	}
	if a.Cookie == "" {
		return errors.New("require cookie")
	}
	return nil
}

func parseArgs() (args, error) {
	a := args{}
	if len(os.Args) >= 2 && !strings.HasPrefix(os.Args[1], "-") {
		filename := os.Args[1]
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return args{}, err
		}
		if err := json.Unmarshal(b, &a); err != nil {
			return args{}, err
		}
		return a, nil
	}
	flag.StringVar(&a.BaseURL, "base-url", "https://kl2-test.kidsloop.net", "base url")
	flag.StringVar(&a.GraphqlEndpoint, "graphql-endpoint", "https://api.beta.kidsloop.net/user/", "graphql endpoint")
	flag.StringVar(&a.OrgID, "org-id", "", "org id, required")
	flag.StringVar(&a.Cookie, "cookie", "", "browser cookie, required")
	scheduleIDsString := flag.String("schedule-ids", "", "schedule ids, separate by comma, require one of \"schedule-ids\" and \"schedule-ids-file\"")
	scheduleIDsFile := flag.String("schedule-ids-file", "", "schedule ids file, separate by newline, require one of \"schedule-ids\" and \"schedule-ids-file\"")
	flag.Parse()
	if scheduleIDsString != nil && *scheduleIDsString != "" {
		items := strings.Split(*scheduleIDsString, ",")
		for _, item := range items {
			item := strings.TrimSpace(item)
			if item != "" {
				a.ScheduleIDs = append(a.ScheduleIDs, item)
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
				a.ScheduleIDs = append(a.ScheduleIDs, item)
			}
		}
	}
	return a, nil
}

func run(a args) error {
	log.Printf("process args: %+v\n", a)
	for _, scheduleID := range a.ScheduleIDs {
		log.Printf("processing: schedule_id = %s ...", scheduleID)
		if err := addAssessment(a.BaseURL, a.GraphqlEndpoint, a.OrgID, scheduleID, a.Cookie); err != nil {
			return err
		}
		log.Printf("process ok: schedule_id = %s", scheduleID)
	}
	return nil
}

func addAssessment(baseURL string, graphqlEndpoint, orgID string, scheduleID string, cookie string) error {
	schedule, err := getSchedule(baseURL, orgID, scheduleID, cookie)
	if err != nil {
		return err
	}
	studentIDs, err := getClassStudentIDs(graphqlEndpoint, schedule.Class.ID, cookie)
	if err != nil {
		return err
	}
	cmd := entity.AddAssessmentArgs{
		ScheduleID:    scheduleID,
		AttendanceIDs: studentIDs,
		ClassLength:   int(time.Unix(schedule.EndAt, 0).Sub(time.Unix(schedule.StartAt, 0))),
		ClassEndTime:  schedule.EndAt,
	}
	b, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, baseURL+"/v1/assessments?org_id="+orgID, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Add("Cookie", cookie)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b2, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("add assessment failed: %d: %s\n: cmd = %+v, tid = %s",
			resp.StatusCode, string(b2), cmd, resp.Header.Get("x-curr-tid")))
	}
	return nil
}

func getSchedule(baseURL string, orgID string, scheduleID string, cookie string) (*entity.ScheduleDetailsView, error) {
	url := baseURL + "/v1/schedules/" + scheduleID + "?org_id=" + orgID
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Cookie", cookie)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		var result entity.ScheduleDetailsView
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}
		return &result, nil
	} else {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(fmt.Sprintf("get schedule fail: %s: %d: %s", url, resp.StatusCode, string(b)))
	}
}

func getClassStudentIDs(graphqlEndpoint string, classID string, cookie string) ([]string, error) {
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
	_, err := chlorine.NewClient(graphqlEndpoint).Run(context.Background(), req, &res)
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
