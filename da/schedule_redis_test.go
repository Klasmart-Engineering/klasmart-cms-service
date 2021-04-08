package da

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"testing"
)

func Test_conditionHash(t *testing.T) {
	str := `
{
        "IDs": {
            "Strings": null,
            "Valid": false
        },
        "OrgID": {
            "String": "32bfe7ba-d897-4504-955c-8d6b484549c62",
            "Valid": true
        },
        "StartAtGe": {
            "Int64": 0,
            "Valid": false
        },
        "StartAtLt": {
            "Int64": 0,
            "Valid": false
        },
        "EndAtLe": {
            "Int64": 0,
            "Valid": false
        },
        "EndAtLt": {
            "Int64": 0,
            "Valid": false
        },
        "EndAtGe": {
            "Int64": 0,
            "Valid": false
        },
        "StartAndEndRange": null,
        "StartAndEndTimeViewRange": [
            {
                "Int64": 1617206400,
                "Valid": true
            },
            {
                "Int64": 1619798399,
                "Valid": true
            }
        ],
        "LessonPlanID": {
            "String": "",
            "Valid": false
        },
        "LessonPlanIDs": {
            "Strings": null,
            "Valid": false
        },
        "RepeatID": {
            "String": "",
            "Valid": false
        },
        "Status": {
            "String": "",
            "Valid": false
        },
        "SubjectIDs": {
            "Strings": null,
            "Valid": false
        },
        "ProgramIDs": {
            "Strings": null,
            "Valid": false
        },
        "RelationID": {
            "String": "",
            "Valid": false
        },
        "RelationIDs": {
            "Strings": [
                "97e42974-81c9-45bc-aa1a-f25db7de3ae4",
                "6d8c6305-72bd-4235-8e3c-0cf67784d13e",
                "All",
                "6d8c6305-72bd-4235-8e3c-0cf67784d13e",
                "All",
                "6d8c6305-72bd-4235-8e3c-0cf67784d13e",
                "All"
            ],
            "Valid": true
        },
        "RelationSchoolIDs": {
            "Strings": null,
            "Valid": false
        },
        "ClassTypes": {
            "Strings": [
                "OnlineClass",
                "Task"
            ],
            "Valid": true
        },
        "DueToEq": {
            "Int64": 0,
            "Valid": false
        },
        "AnyTime": {
            "Bool": false,
            "Valid": false
        },
        "OrderBy": 3,
        "Pager": {
            "Page": 0,
            "PageSize": 0
        },
        "DeleteAt": {
            "Int64": 0,
            "Valid": false
        }
    }`
	condition := new(ScheduleCondition)
	json.Unmarshal([]byte(str), condition)
	//t.Log(condition)
	key := conditionHash(ScheduleCacheCondition{Condition: condition})

	t.Log(key)
}

func conditionHash2(condition dbo.Conditions) string {
	h := md5.New()
	str := fmt.Sprintf("%v", condition)
	h.Write([]byte(str))
	md5Hash := fmt.Sprintf("%x", h.Sum(nil))
	fmt.Println(str)
	return md5Hash
}

func conditionHash(condition ScheduleCacheCondition) string {
	h := md5.New()
	str := fmt.Sprintf("%v", condition)
	h.Write([]byte(str))
	md5Hash := fmt.Sprintf("%x", h.Sum(nil))
	fmt.Println(str)
	return md5Hash
}
