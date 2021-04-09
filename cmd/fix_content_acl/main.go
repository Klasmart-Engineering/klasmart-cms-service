package main

import (
	"context"
	"flag"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
)

const (
	granteeGroupURIAllUser = "http://acs.amazonaws.com/groups/global/AllUsers"
	defaultGranteeID       = "e290d6b16a91aa0310f5d5a54eb01f340486ff1e8b96e2a01ac9dad4ad9c32f8"
)

var (
	bucket          string
	region          string
	accessKeyID     string
	secretAccessKey string
	prefix          string
	granteeID       string
	workerCount     int
)

func init() {
	flag.StringVar(&bucket, "b", "", "Bucket name,required.")
	flag.StringVar(&region, "r", "", "Region name,required.")
	flag.StringVar(&accessKeyID, "a", "", "S3 access key ID,required.")
	flag.StringVar(&secretAccessKey, "s", "", "S3 secret access key,required.")
	flag.StringVar(&prefix, "p", "", "Query objects prefix.")
	flag.StringVar(&granteeID, "g", "", "Grantee AWS account ID.")
	flag.IntVar(&workerCount, "w", 0, "Worker pool size,default is runtime.NumCPU().")

	if granteeID == "" {
		granteeID = defaultGranteeID
	}

	if workerCount < 1 {
		workerCount = runtime.NumCPU()
	}
}

func main() {
	flag.Parse()

	ctx := context.Background()

	log.Info(ctx, "Arguments",
		log.String("Region", region),
		log.String("Bucket", bucket),
		log.String("AccessKeyID", accessKeyID),
		log.String("SecretAccessKey", secretAccessKey),
		log.String("Prefix", prefix),
		log.String("GranteeID", granteeID),
		log.Int("WorkerCount", workerCount))

	svc, err := NewS3Client(ctx, accessKeyID, secretAccessKey, region)
	if err != nil {
		return
	}

	objKeys, err := ListObjectKeys(ctx, svc, bucket, prefix)
	if err != nil {
		return
	}

	log.Info(ctx, "Object keys.",
		log.String("Bucket:", bucket),
		log.String("Prefix:", prefix),
		log.Int("Key count:", len(objKeys)))

	var publicAccessObjCount uint32

	jobs := make(chan int, 100)
	wg := sync.WaitGroup{}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				objKey := objKeys[j]
				objAclOut, err := GetObjectACL(ctx, svc, bucket, objKey)
				if err != nil {
					continue
				}

				for k := range objAclOut.Grants {
					if !IsPublicAccess(objAclOut.Grants[k]) {
						continue
					}

					log.Info(ctx, "Put object acl", log.String("Key", objKey))
					atomic.AddUint32(&publicAccessObjCount, 1)
					err = PutObjectACL(ctx, svc, bucket, objKey, granteeID)
					if err != nil {
						continue
					}
					break
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for i := range objKeys {
			jobs <- i
		}
	}()

	wg.Wait()

	log.Info(ctx, "Done.",
		log.Uint32("Public access count", publicAccessObjCount))
}

func NewS3Client(ctx context.Context, accessKeyID, secretAccessKey, region string) (s3iface.S3API, error) {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		Region:      aws.String(region),
	})
	if err != nil {
		log.Error(ctx, "Create AWS session failed",
			log.Err(err),
			log.String("accessKeyID", accessKeyID),
			log.String("secretAccessKey", secretAccessKey),
			log.String("region", region),
		)
		return nil, err
	}

	svc := s3.New(sess)
	return svc, nil
}

// GetObjectACL gets the ACL for a bucket object
func GetObjectACL(ctx context.Context, svc s3iface.S3API, bucket, key string) (*s3.GetObjectAclOutput, error) {
	result, err := svc.GetObjectAcl(&s3.GetObjectAclInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Error(ctx, "Get object acl failed",
			log.Err(err),
			log.String("bucket", bucket),
			log.String("key", key))
		return nil, err
	}
	return result, nil
}

// PutObjectACL gives the person with AWS account user ID access to BUCKET OBJECT
func PutObjectACL(ctx context.Context, svc s3iface.S3API, bucket, key, ID string) error {
	granteeIDStr := aws.String("id=" + ID)
	// Default config READ WRITE READ_ACP WRITE_ACP to BUCKET OBJECT
	_, err := svc.PutObjectAcl(&s3.PutObjectAclInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		GrantRead:     granteeIDStr,
		GrantWrite:    granteeIDStr,
		GrantReadACP:  granteeIDStr,
		GrantWriteACP: granteeIDStr,
	})
	if err != nil {
		log.Error(ctx, "Put object acl failed",
			log.Err(err),
			log.String("bucket", bucket),
			log.String("key", key),
			log.String("GranteeID", ID))

		return err
	}

	return nil
}

// ListObjectsKeys iterate all objects in an Amazon S3 bucket
func ListObjectKeys(ctx context.Context, svc s3iface.S3API, bucket, prefix string) ([]string, error) {
	objKeys := []string{}
	input := &s3.ListObjectsV2Input{Bucket: aws.String(bucket)}
	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	output, err := svc.ListObjectsV2(input)
	if err != nil {
		log.Error(ctx, "List objects failed",
			log.Err(err),
			log.String("bucket", bucket),
			log.String("prefix", prefix))
		return nil, err
	}

	for i := range output.Contents {
		objKeys = append(objKeys, aws.StringValue(output.Contents[i].Key))
	}

	for output.NextContinuationToken != nil {
		input.SetContinuationToken(*output.NextContinuationToken)

		output, err = svc.ListObjectsV2(input)
		if err != nil {
			log.Error(ctx, "List objects failed",
				log.Err(err),
				log.String("bucket", bucket),
				log.String("prefix", prefix),
				log.String("ContinuationToken", *input.ContinuationToken))
			return nil, err
		}

		for i := range output.Contents {
			objKeys = append(objKeys, aws.StringValue(output.Contents[i].Key))
		}
	}

	return objKeys, nil
}

func IsPublicAccess(grand *s3.Grant) bool {
	if aws.StringValue(grand.Grantee.Type) == s3.TypeGroup && aws.StringValue(grand.Grantee.URI) == granteeGroupURIAllUser {
		return true
	}

	return false
}
