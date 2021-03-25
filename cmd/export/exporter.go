package main

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"io"
	"os"
	"strings"

	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/model/storage"
)

const (
	AssetsTagPath        = "assets"
	TeacherManualTagPath = "teacher_manual"
	ThumbnailTagPath     = "thumbnail"
)

type ContentExporter struct {
}

type ResourceBatch struct {
	Assets        []string
	TeacherManual []string
	Thumbnail     []string
}

type ResourceObject struct {
	Partition storage.StoragePartition
	Name      string
	Path      string
}

//Export run export process
func (c *ContentExporter) Export(ctx context.Context, condition da.ContentCondition, path string) error {
	contentList, err := c.GetExportContents(ctx, condition)
	if err != nil {
		fmt.Printf("Get export contents failed, err: %v\n", err)
		return err
	}
	batch, err := c.CollectRelatedResources(ctx, contentList)
	if err != nil {
		fmt.Printf("Collect related resources failed, err: %v\n", err)
		return err
	}
	err = c.DownloadRelatedResources(ctx, batch, path)
	if err != nil {
		fmt.Printf("Download related resources failed, err: %v\n", err)
		return err
	}
	return nil
}

//DownloadRelatedResources download resources
func (c *ContentExporter) DownloadRelatedResources(ctx context.Context, resourceBatch *ResourceBatch, path string) error {
	resourceObjects := make([]*ResourceObject, 0)
	//Collect assets resource objects
	for i := range resourceBatch.Assets {
		partition, name, err := c.parseResource(ctx, resourceBatch.Assets[i])
		if err != nil {
			fmt.Printf("Parse resource failed, err: %v\n", err)
			return err
		}
		resourceObjects = append(resourceObjects, &ResourceObject{
			Partition: partition,
			Name:      name,
			Path:      path + "/" + AssetsTagPath + "/" + name,
		})
	}
	//Collect thumbnail resource objects
	for i := range resourceBatch.Thumbnail {
		partition, name, err := c.parseResource(ctx, resourceBatch.Assets[i])
		if err != nil {
			fmt.Printf("Parse resource failed, err: %v\n", err)
			return err
		}
		resourceObjects = append(resourceObjects, &ResourceObject{
			Partition: partition,
			Name:      name,
			Path:      path + "/" + ThumbnailTagPath + "/" + name,
		})
	}
	//Collect teachermanual resource objects
	for i := range resourceBatch.TeacherManual {
		partition, name, err := c.parseResource(ctx, resourceBatch.Assets[i])
		if err != nil {
			fmt.Printf("Parse resource failed, err: %v\n", err)
			return err
		}
		resourceObjects = append(resourceObjects, &ResourceObject{
			Partition: partition,
			Name:      name,
			Path:      path + "/" + TeacherManualTagPath + "/" + name,
		})
	}

	for i := range resourceObjects {
		reader, err := storage.DefaultStorage().DownloadFile(ctx, resourceObjects[i].Partition, resourceObjects[i].Name)
		if err != nil {
			fmt.Printf("Download resource failed, err: %v\n", err)
			return err
		}
		f, err := os.Create(resourceObjects[i].Path)
		if err != nil {
			fmt.Printf("Create dist file failed, err: %v\n", err)
			return err
		}
		_, err = io.Copy(f, reader)
		if err != nil {
			fmt.Printf("Save Resource failed, err: %v\n", err)
			return err
		}
	}

	return nil
}

//CollectRelatedResources collect all related resources from content list to download
func (c *ContentExporter) CollectRelatedResources(ctx context.Context, contentList []*entity.Content) (*ResourceBatch, error) {
	ret := new(ResourceBatch)
	for i := range contentList {
		//Add thumbnail
		ret.Thumbnail = append(ret.Thumbnail, contentList[i].Thumbnail)

		contentData, err := model.CreateContentData(ctx, contentList[i].ContentType, contentList[i].Data)
		if err != nil {
			fmt.Printf("Can't unmarshal content plan data, id: %v, data: %v, err: %v\n", contentList[i].ID, contentList[i].Data, err)
			return nil, err
		}
		switch v := contentData.(type) {
		case *model.AssetsData:
			//Add assets
			ret.Assets = append(ret.Assets, string(v.Source))
		case *model.LessonData:
			//Add teacher manual if the plan has teacher manual
			if v.TeacherManual != "" {
				ret.TeacherManual = append(ret.TeacherManual, string(v.TeacherManual))
			}
		case *model.MaterialData:
			//Add the material assets, if the material is h5p, ignore
			if v.FileType != entity.FileTypeH5p {
				ret.Assets = append(ret.Assets, string(v.Source))
			}
		}
	}
	return ret, nil
}

//GetExportContents get all contents needed to export
func (c *ContentExporter) GetExportContents(ctx context.Context, condition da.ContentCondition) ([]*entity.Content, error) {
	//Get immediate contents
	contentList, relatedIDs, err := c.GetExportImmediateContents(ctx, condition)
	if err != nil {
		fmt.Printf("Can't get exported contents, err: %v\n", err)
		return nil, err
	}
	//Get related contents
	relatedContentList, err := c.GetExportRelatedMaterials(ctx, relatedIDs)
	if err != nil {
		fmt.Printf("Can't get related contents, err: %v\n", err)
		return nil, err
	}
	contentList = append(contentList, relatedContentList...)
	return contentList, nil
}

//GetExportImmediateContents get content list to export by search condition
func (c *ContentExporter) GetExportImmediateContents(ctx context.Context, condition da.ContentCondition) ([]*entity.Content, []string, error) {
	//Search content by conditions
	_, contentList, err := da.GetContentDA().SearchContent(ctx, dbo.MustGetDB(ctx), condition)
	if err != nil {
		fmt.Printf("Can't get contents, err: %v\n", err)
		return nil, nil, err
	}
	//Get materials from contents
	relatedIDs := make([]string, 0)
	for i := range contentList {
		//if content is not plan, pass
		//get related material ids from plans
		if contentList[i].ContentType != entity.ContentTypePlan {
			continue
		}
		planData, err := model.CreateContentData(ctx, contentList[i].ContentType, contentList[i].Data)
		if err != nil {
			fmt.Printf("Can't unmarshal content plan data, id: %v, data: %v, err: %v\n", contentList[i].ID, contentList[i].Data, err)
			return nil, nil, err
		}
		err = planData.PrepareVersion(ctx)
		if err != nil {
			fmt.Printf("Can't prepare version for content plan data, id: %v, data: %v, err: %v\n", contentList[i].ID, contentList[i].Data, err)
			return nil, nil, err
		}
		relatedIDs = append(relatedIDs, planData.SubContentIDs(ctx)...)
	}
	return contentList, relatedIDs, nil
}

//GetExportRelatedMaterials get related materials list to export by exported plans
func (c *ContentExporter) GetExportRelatedMaterials(ctx context.Context, requiredExtIDs []string) ([]*entity.Content, error) {
	_, contentList, err := da.GetContentDA().SearchContent(ctx, dbo.MustGetDB(ctx), da.ContentCondition{
		IDS: requiredExtIDs,
	})
	if err != nil {
		fmt.Printf("Can't get contents, err: %v\n", err)
		return nil, err
	}
	return contentList, nil
}

func (c *ContentExporter) parseResource(ctx context.Context, resource string) (storage.StoragePartition, string, error) {
	resourcePairs := strings.Split(resource, constant.TeacherManualSeparator)
	if len(resourcePairs) != 2 {
		fmt.Printf("Invalid resource id, resource: %v\n", resourcePairs)
		return "", "", entity.ErrInvalidResourceId
	}
	extensionPairs := strings.Split(resourcePairs[1], ".")
	if len(extensionPairs) != 2 {
		log.Error(ctx, "invalid extension", log.String("resourceId", resource))
		return "", "", entity.ErrInvalidResourceId
	}
	partition, err := storage.NewStoragePartition(ctx, resourcePairs[0], extensionPairs[1])
	if err != nil {
		fmt.Printf("Invalid partition name, resource: %v\n", resourcePairs)
		return "", "", err
	}
	return partition, resourcePairs[1], nil
}
