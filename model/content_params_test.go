package model

import (
	"context"
	"encoding/json"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"testing"
)

func TestUnmarshalContentData(t *testing.T) {
	data := "{\"segmentId\":\"1\",\"condition\":\"\",\"materialId\":\"5fbdb67de0ad99a97be8681c\",\"material\":null,\"next\":[{\"segmentId\":\"11\",\"condition\":\"\",\"materialId\":\"5fbcd442869fc4f88ee362cd\",\"material\":null,\"next\":[{\"segmentId\":\"111\",\"condition\":\"\",\"materialId\":\"5fbcbc5f4a60a5856c1abbea\",\"material\":null,\"next\":[],\"teacher_manual\":\"\"}],\"teacher_manual\":\"\"}],\"teacher_manual\":\"teacher_manual-5fbf0d0c25b6981b872b7988.pdf\"}"
	ins := LessonData{}
	err := json.Unmarshal([]byte(data), &ins)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("Done")
}

func TestMigrateInvalidData(t *testing.T) {
	ctx := context.Background()
	total, contents, err := da.GetContentDA().SearchContent(ctx, dbo.MustGetDB(ctx), da.ContentCondition{})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(total)
	count := 0
	for i := range contents{
		if contents[i].ContentType == entity.ContentTypeMaterial {
			contentData, err := CreateContentData(ctx, entity.ContentTypeMaterial, contents[i].Data)
			if err != nil{
				t.Error(err)
				return
			}
			materialData := contentData.(*MaterialData)
			if materialData.FileType == 0 || materialData.InputSource == 0 {
				switch materialData.Source.Ext() {
					case "jpg":
						fallthrough
					case "png":
						materialData.FileType = entity.FileTypeImage
						materialData.InputSource = entity.MaterialInputSourceDisk
					case "mp4":
						materialData.FileType = entity.FileTypeVideo
						materialData.InputSource = entity.MaterialInputSourceDisk
					case "mp3":
						materialData.FileType = entity.FileTypeAudio
						materialData.InputSource = entity.MaterialInputSourceDisk
				}
				data, err := materialData.Marshal(ctx)
				if err != nil{
					t.Error(err)
					return
				}
				contents[i].Data = data
				err = da.GetContentDA().UpdateContent(ctx, dbo.MustGetDB(ctx), contents[i].ID, *contents[i])
				if err != nil {
					t.Error(err)
					return
				}

				//t.Logf("data: %v, material: %#v", contents[i].Data, contents[i])
				count ++
			}

		}
	}
	t.Log(count)
}

func TestMigrateContentData(t *testing.T) {
	Author := "b2275a92-06cd-5837-b127-850ceeb03c8a"
	Org := "8901cdbe-3e6e-46b2-ad2a-bb1927a5fbc4"

	ctx := context.Background()
	err := dbo.MustGetDB(ctx).AutoMigrate(entity.Content{}).Error
	if err != nil{
		t.Error(err)
		return
	}
	total, contents, err := da.GetContentDA().SearchContent(ctx, dbo.MustGetDB(ctx), da.ContentCondition{})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(total)
	count := 0
	for i := range contents {
		data := contents[i].Data
		switch contents[i].ContentType {
		case 10:
			tempContent, err := migrateContentData(ctx, contents[i], entity.FileTypeImage)
			if err != nil {
				t.Error(err)
				return
			}
			contents[i] = tempContent
		case 11:
			tempContent, err := migrateContentData(ctx, contents[i], entity.FileTypeVideo)
			if err != nil {
				t.Error(err)
				return
			}
			contents[i] = tempContent
		case 12:
			tempContent, err := migrateContentData(ctx, contents[i], entity.FileTypeAudio)
			if err != nil {
				t.Error(err)
				return
			}
			contents[i] = tempContent
		case entity.ContentTypeMaterial:
			contentData, err := CreateContentData(ctx, entity.ContentTypeMaterial, contents[i].Data)
			if err != nil {
				t.Error(err)
				return
			}
			materialData := contentData.(*MaterialData)

			if materialData.FileType == 0 || materialData.InputSource == 0  {
				switch materialData.Source.Ext() {
				case "jpg":
					fallthrough
				case "png":
					materialData.FileType = entity.FileTypeImage
					materialData.InputSource = entity.MaterialInputSourceDisk
				case "mp4":
					materialData.FileType = entity.FileTypeVideo
					materialData.InputSource = entity.MaterialInputSourceDisk
				case "mp3":
					materialData.FileType = entity.FileTypeAudio
					materialData.InputSource = entity.MaterialInputSourceDisk
				case "":
					materialData.FileType = 5
					materialData.InputSource = entity.MaterialInputSourceH5p
				}
				data, err = materialData.Marshal(ctx)
				if err != nil {
					t.Error(err)
					return
				}
			}
		}
		contents[i].Data = data
		contents[i].Author = Author
		contents[i].Org = Org
		contents[i].PublishScope = Org
		contents[i].Creator = Author
		contents[i].DirPath = constant.FolderRootPath
		if contents[i].Grade == "grade1" {
			contents[i].Grade = "grade0"
		}

		err = da.GetContentDA().UpdateContent(ctx, dbo.MustGetDB(ctx), contents[i].ID, *contents[i])
		if err != nil {
			t.Error(err)
			return
		}

	}
	t.Log(count)
}

func migrateContentData(ctx context.Context, content *entity.Content, fileType entity.FileType) (*entity.Content, error){
	content.ContentType = entity.ContentTypeMaterial
	contentData, err := CreateContentData(ctx, entity.ContentTypeMaterial, content.Data)
	if err != nil {
		return nil, err
	}
	materialData := contentData.(*MaterialData)
	materialData.FileType = fileType
	materialData.InputSource = entity.MaterialInputSourceDisk
	data, err := materialData.Marshal(ctx)
	if err != nil {
		return nil, err
	}
	content.Data = data
	return content, nil
}