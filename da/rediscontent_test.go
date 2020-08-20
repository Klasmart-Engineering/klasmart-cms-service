package da

import (
	"context"
	"github.com/go-redis/redis"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"testing"
)

func TestRedisCreateContent(t *testing.T){
	ro.SetConfig(&redis.Options{
		Addr:               "172.20.219.158:6379",
		Password:           "",
	})
	cid, err := GetRedisContentDA().CreateContent(context.Background(), entity.Content{
		ContentType:   1,
		Name:          "张三爱科学",
		Keywords:      "逻辑,计算机",
		Description:   "很好的学习内容",
		Data:          "123",
		Extra:         "555",
		Author:        "zhangsan",
		AuthorName:    "张三",
		Org:           "张三出版社",
		PublishScope:  "123",
		PublishStatus: "published",
		Version:       1,
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(cid)

	data, err := GetRedisContentDA().GetContentById(context.Background(), cid)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Content: %#v\n", data)
}



func TestRedisCreateContent2(t *testing.T){
	ro.SetConfig(&redis.Options{
		Addr:               "172.20.219.158:6379",
		Password:           "",
	})
	cid, err := GetRedisContentDA().CreateContent(context.Background(), entity.Content{
		ContentType:   1,
		Name:          "张三爱科学",
		Keywords:      "逻辑,计算机",
		Description:   "很好的学习内容",
		Data:          "123",
		Extra:         "555",
		Author:        "zhangsanfeng",
		AuthorName:    "张三",
		Org:           "张三出版社",
		PublishScope:  "123",
		PublishStatus: "published",
		Version:       1,
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(cid)

	data, err := GetRedisContentDA().GetContentById(context.Background(), cid)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Content: %#v\n", data)
	err = GetRedisContentDA().DeleteContent(context.Background(), cid)
	if err != nil {
		t.Error(err)
		return
	}

	dataList, err := GetRedisContentDA().SearchContent(context.Background(), RedisContentCondition{
		KeyWords:      "张三",
	})
	if err != nil {
		t.Error(err)
		return
	}
	for i := range dataList{
		t.Logf("Content: %#v\n", dataList[i])
	}
}