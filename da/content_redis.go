package da

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"gitlab.badanamu.com.cn/calmisland/ro"
	"strconv"
	"sync"
)

const(
	RedisContentBucket = "kidsloop2:da:content"
	RedisContentKeywordsBucket = "kidsloop2:da:content:keywords"
	RedisContentContentTypeBucket = "kidsloop2:da:content:contentType"
	RedisContentPublishStatusBucket = "kidsloop2:da:content:publishStatus"
	RedisContentAuthorBucket = "kidsloop2:da:content:author"
	RedisContentOrgBucket = "kidsloop2:da:content:org"
)

type RedisContentDA struct {

}

func (r *RedisContentDA) setIndex(ctx context.Context, co entity.Content) error {
	err := ro.MustGetRedis(ctx).SAdd(redisKey(ctx, RedisContentKeywordsBucket, co.Name), co.ID).Err()
	if err != nil{
		return err
	}
	err = ro.MustGetRedis(ctx).SAdd(redisKey(ctx, RedisContentKeywordsBucket, co.Keywords), co.ID).Err()
	if err != nil{
		return err
	}
	err = ro.MustGetRedis(ctx).SAdd(redisKey(ctx, RedisContentKeywordsBucket, co.Description), co.ID).Err()
	if err != nil{
		return err
	}
	err = ro.MustGetRedis(ctx).SAdd(redisKey(ctx, RedisContentKeywordsBucket, co.AuthorName), co.ID, ).Err()
	if err != nil{
		return err
	}
	err = ro.MustGetRedis(ctx).SAdd(redisKey(ctx, RedisContentContentTypeBucket, strconv.Itoa(int(co.ContentType))), co.ID).Err()
	if err != nil{
		return err
	}
	err = ro.MustGetRedis(ctx).SAdd(redisKey(ctx, RedisContentPublishStatusBucket, string(co.PublishStatus)), co.ID).Err()
	if err != nil{
		return err
	}
	err = ro.MustGetRedis(ctx).SAdd(redisKey(ctx, RedisContentAuthorBucket, co.AuthorName), co.ID).Err()
	if err != nil{
		return err
	}
	err = ro.MustGetRedis(ctx).SAdd(redisKey(ctx, RedisContentOrgBucket, co.Org), co.ID).Err()
	if err != nil{
		return err
	}
	return nil
}

func (r *RedisContentDA) removeIndex(ctx context.Context, co entity.Content) error {
	err := ro.MustGetRedis(ctx).SRem(redisKey(ctx, RedisContentKeywordsBucket, co.Name), co.ID).Err()
	if err != nil{
		return err
	}
	err = ro.MustGetRedis(ctx).SRem(redisKey(ctx, RedisContentKeywordsBucket, co.Keywords), co.ID).Err()
	if err != nil{
		return err
	}
	err = ro.MustGetRedis(ctx).SRem(redisKey(ctx, RedisContentKeywordsBucket, co.Description), co.ID).Err()
	if err != nil{
		return err
	}
	err = ro.MustGetRedis(ctx).SRem(redisKey(ctx, RedisContentKeywordsBucket, co.AuthorName), co.ID).Err()
	if err != nil{
		return err
	}
	err = ro.MustGetRedis(ctx).SRem(redisKey(ctx, RedisContentContentTypeBucket, strconv.Itoa(int(co.ContentType))), co.ID).Err()
	if err != nil{
		return err
	}
	err = ro.MustGetRedis(ctx).SRem(redisKey(ctx, RedisContentPublishStatusBucket, string(co.PublishStatus)), co.ID).Err()
	if err != nil{
		return err
	}
	err = ro.MustGetRedis(ctx).SRem(redisKey(ctx, RedisContentAuthorBucket, co.Author), co.ID).Err()
	if err != nil{
		return err
	}
	err = ro.MustGetRedis(ctx).SRem(redisKey(ctx, RedisContentOrgBucket, co.Org), co.ID).Err()
	if err != nil{
		return err
	}
	return nil
}

func (r *RedisContentDA) CreateContent(ctx context.Context, co entity.Content) (string, error) {
	co.ID = utils.NewID()
	data, err := json.Marshal(co)
	if err != nil{
		return "", err
	}
	err = ro.MustGetRedis(ctx).SetNX(redisKey(ctx, RedisContentBucket, co.ID), data, 0).Err()
	if err != nil{
		return "", err
	}

	err = r.setIndex(ctx, co)
	if err != nil{
		return "", err
	}

	return co.ID, nil
}

func (r *RedisContentDA) UpdateContent(ctx context.Context, cid string, co entity.Content) error {
	oldContent, err := r.GetContentById(ctx, cid)
	if err != nil{
		return err
	}
	err = r.removeIndex(ctx, *oldContent)
	if err != nil{
		return err
	}
	data, err := json.Marshal(co)
	if err != nil{
		return err
	}
	err = ro.MustGetRedis(ctx).Set(redisKey(ctx, RedisContentBucket, co.ID), data, -1).Err()
	if err != nil{
		return err
	}
	err = r.setIndex(ctx, co)
	if err != nil{
		return err
	}
	return nil
}

func (r *RedisContentDA) DeleteContent(ctx context.Context, cid string) error {
	oldContent, err := r.GetContentById(ctx, cid)
	if err != nil{
		return err
	}
	err = r.removeIndex(ctx, *oldContent)
	if err != nil{
		return err
	}
	err = ro.MustGetRedis(ctx).Del(redisKey(ctx, RedisContentBucket, cid)).Err()
	if err != nil{
		return err
	}
	return nil
}

func (r *RedisContentDA) GetContentById(ctx context.Context, cid string) (*entity.Content, error) {
	res, err := ro.MustGetRedis(ctx).Get(redisKey(ctx, RedisContentBucket, cid)).Result()
	if err != nil {
		return nil, err
	}
	content := new(entity.Content)
	err = json.Unmarshal([]byte(res), content)
	if err != nil{
		return nil, err
	}
	return content, nil
}

func (r *RedisContentDA) SearchContent(ctx context.Context, condition RedisContentCondition) ([]*entity.Content, error) {
	ids := condition.GetConditions(ctx)
	if len(ids) == 0 {
		return nil, nil
	}
	keys := make([]string, len(ids))
	for i := range ids {
		keys[i] = redisKey(ctx, RedisContentBucket, ids[i])
	}

	res, err := ro.MustGetRedis(ctx).MGet(keys...).Result()
	if err != nil{
		return nil, err
	}

	contentList := make([]*entity.Content, len(res))
	for i := range res {
		resData, ok := res[i].(string)
		if !ok {
			continue
		}
		tempContent := new(entity.Content)
		err = json.Unmarshal([]byte(resData), tempContent)
		if err != nil{
			return nil, err
		}
		contentList[i] = tempContent
	}
	return contentList, nil
}

type RedisContentCondition struct {
	IDS          []string `json:"ids"`
	KeyWords     string `json:"keywords"`		//name,keyword,authorName,description
	ContentType  []int `json:"content_type"`
	PublishStatus []string `json:"publish_status"`
	AuthorName      string `json:"author"`
	Org    		string `json:"org"`
}

func (r *RedisContentCondition) GetConditions(ctx context.Context) []string {
	ids := make([]string, 0)
	if r.IDS != nil {
		ids = append(ids, r.IDS...)
	}
	if r.KeyWords != "" {
		key := redisKey(ctx, RedisContentKeywordsBucket, r.KeyWords)
		ids = append(ids, ro.MustGetRedis(ctx).SMembers(key).Val()...)
	}
	if r.ContentType != nil {
		for i := range r.ContentType {
			key := redisKey(ctx, RedisContentContentTypeBucket, strconv.Itoa(r.ContentType[i]))
			ids = append(ids, ro.MustGetRedis(ctx).SMembers(key).Val()...)
		}
	}
	if r.PublishStatus != nil {
		for i := range r.PublishStatus {
			key := redisKey(ctx, RedisContentPublishStatusBucket, r.PublishStatus[i])
			ids = append(ids, ro.MustGetRedis(ctx).SMembers(key).Val()...)
		}
	}
	if r.AuthorName != "" {
		key := redisKey(ctx, RedisContentAuthorBucket, r.AuthorName)
		ids = append(ids, key)
	}
	if r.Org != "" {
		key := redisKey(ctx, RedisContentOrgBucket, r.Org)
		ids = append(ids, key)
	}
	return ids
}

func redisKey(ctx context.Context, bucket, key string) string {
	return fmt.Sprintf("%v:%v", bucket, key)
}

var(
	_redisContentDA *RedisContentDA
	_redisContentDAOnce sync.Once
)

func GetRedisContentDA() *RedisContentDA {
	_redisContentDAOnce.Do(func() {
		_redisContentDA = new(RedisContentDA)
	})
	return _redisContentDA
}
