package model

import (
	"context"
	"database/sql"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/dbo"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/da"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"sync"
	"time"
)

type IProgramModel interface {
	Query(ctx context.Context, condition *da.ProgramCondition) ([]*entity.Program, error)
	GetByID(ctx context.Context, id string) (*entity.Program, error)
	Add(ctx context.Context, op *entity.Operator, data *entity.Program) (string, error)
	Update(ctx context.Context, op *entity.Operator, data *entity.Program) (string, error)
	Delete(ctx context.Context, op *entity.Operator, id string) error
	SetAge(ctx context.Context, id string, ageIDs []string) error
	SetGrade(ctx context.Context, id string, gradeIDs []string) error
	SetSubject(ctx context.Context, id string, subjectIDs []string) error
	SetDevelopment(ctx context.Context, id string, developmentIDs []string) error
	SetDeveSkill(ctx context.Context, id string, developmentID string, skillIDs []string) error
}

type programModel struct {
}

func (m *programModel) Delete(ctx context.Context, op *entity.Operator, id string) error {
	var old = new(entity.Program)
	err := da.GetProgramDA().Get(ctx, id, old)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "record not found", log.Err(err), log.String("id", id))
		return nil
	}
	if err != nil {
		log.Error(ctx, "get error", log.Err(err), log.String("id", id))
		return err
	}
	if old.DeleteAt != 0 {
		log.Error(ctx, "record is deleted", log.Err(err), log.String("id", id), log.Any("old", old))
		return nil
	}
	old.DeleteAt = time.Now().Unix()
	old.DeleteID = op.UserID
	_, err = da.GetProgramDA().Update(ctx, old)
	if err != nil {
		log.Error(ctx, "update error", log.Err(err), log.String("id", id), log.Any("old", old))
		return err
	}
	return nil
}

func (m *programModel) SetDeveSkill(ctx context.Context, id string, developmentID string, skillIDs []string) error {
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		// remove
		err := da.GetProgramDeveSkillDA().DeleteByProgramIDAndDevelopmentID(ctx, tx, id, developmentID)
		if err != nil {
			log.Error(ctx, "delete error", log.Err(err), log.Any("programID", id))
			return err
		}
		for _, skillID := range skillIDs {
			item := &entity.DevelopmentSkill{
				ID:            utils.NewID(),
				ProgramID:     id,
				DevelopmentID: developmentID,
				SkillID:       skillID,
			}
			_, err := da.GetProgramDeveSkillDA().InsertTx(ctx, tx, item)
			if err != nil {
				log.Error(ctx, "add error", log.Err(err), log.Any("item", item))
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Error(ctx, "SetGrade error", log.Err(err), log.String("programID", id), log.String("developmentID", developmentID), log.Strings("skillIDs", skillIDs))
		return err
	}
	return nil
}

func (m *programModel) SetDevelopment(ctx context.Context, id string, developmentIDs []string) error {
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		// remove
		err := da.GetProgramDevelopmentDA().DeleteByProgramID(ctx, tx, id)
		if err != nil {
			log.Error(ctx, "delete error", log.Err(err), log.Any("programID", id))
			return err
		}
		// set
		for _, developmentID := range developmentIDs {
			item := &entity.ProgramDevelopment{
				ID:            utils.NewID(),
				ProgramID:     id,
				DevelopmentID: developmentID,
			}
			_, err := da.GetProgramDevelopmentDA().InsertTx(ctx, tx, item)
			if err != nil {
				log.Error(ctx, "add error", log.Err(err), log.Any("item", item))
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Error(ctx, "SetGrade error", log.Err(err), log.String("programID", id), log.Strings("developmentIDs", developmentIDs))
		return err
	}
	return nil
}

func (m *programModel) SetSubject(ctx context.Context, id string, subjectIDs []string) error {
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		// remove
		err := da.GetProgramSubjectDA().DeleteByProgramID(ctx, tx, id)
		if err != nil {
			log.Error(ctx, "delete error", log.Err(err), log.Any("programID", id))
			return err
		}
		// set
		for _, subjectID := range subjectIDs {
			item := &entity.ProgramSubject{
				ID:        utils.NewID(),
				ProgramID: id,
				SubjectID: subjectID,
			}
			_, err := da.GetProgramSubjectDA().InsertTx(ctx, tx, item)
			if err != nil {
				log.Error(ctx, "add error", log.Err(err), log.Any("item", item))
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Error(ctx, "SetGrade error", log.Err(err), log.String("programID", id), log.Strings("subjectIDs", subjectIDs))
		return err
	}
	return nil
}
func (m *programModel) SetGrade(ctx context.Context, id string, gradeIDs []string) error {
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		// remove
		err := da.GetProgramGradeDA().DeleteByProgramID(ctx, tx, id)
		if err != nil {
			log.Error(ctx, "delete error", log.Err(err), log.Any("programID", id))
			return err
		}
		// set
		for _, gradeID := range gradeIDs {
			item := &entity.ProgramGrade{
				ID:        utils.NewID(),
				ProgramID: id,
				GradeID:   gradeID,
			}
			_, err := da.GetProgramGradeDA().InsertTx(ctx, tx, item)
			if err != nil {
				log.Error(ctx, "add error", log.Err(err), log.Any("item", item))
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Error(ctx, "SetGrade error", log.Err(err), log.String("programID", id), log.Strings("gradeIDs", gradeIDs))
		return err
	}
	return nil
}

func (m *programModel) SetAge(ctx context.Context, id string, ageIDs []string) error {
	err := dbo.GetTrans(ctx, func(ctx context.Context, tx *dbo.DBContext) error {
		// remove
		err := da.GetProgramAgeDA().DeleteByProgramID(ctx, tx, id)
		if err != nil {
			log.Error(ctx, "delete error", log.Err(err), log.Any("programID", id))
			return err
		}
		// set
		for _, ageID := range ageIDs {
			item := &entity.ProgramAge{
				ID:        utils.NewID(),
				ProgramID: id,
				AgeID:     ageID,
			}
			_, err := da.GetProgramAgeDA().InsertTx(ctx, tx, item)
			if err != nil {
				log.Error(ctx, "add error", log.Err(err), log.Any("item", item))
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Error(ctx, "SetAge error", log.Err(err), log.String("programID", id), log.Strings("ageIDs", ageIDs))
		return err
	}
	return nil
}

func (m *programModel) Add(ctx context.Context, op *entity.Operator, data *entity.Program) (string, error) {
	data.ID = utils.NewID()
	data.CreateAt = time.Now().Unix()
	data.CreateID = op.UserID
	_, err := da.GetProgramDA().Insert(ctx, data)
	if err != nil {
		log.Error(ctx, "add error", log.Err(err), log.Any("data", data))
		return "", err
	}
	return data.ID, nil
}

func (m *programModel) Update(ctx context.Context, op *entity.Operator, data *entity.Program) (string, error) {
	var old = new(entity.Program)
	err := da.GetProgramDA().Get(ctx, data.ID, old)
	if err != nil {
		log.Error(ctx, "get error", log.Err(err), log.Any("data", data))
		return "", err
	}
	old.Name = data.Name
	old.Number = data.Number
	old.OrgType = data.OrgType
	old.UpdateID = op.UserID
	old.UpdateAt = time.Now().Unix()
	_, err = da.GetProgramDA().Update(ctx, old)
	if err != nil {
		log.Error(ctx, "update error", log.Err(err), log.Any("data", data))
		return "", err
	}
	return old.ID, nil
}

func (m *programModel) Query(ctx context.Context, condition *da.ProgramCondition) ([]*entity.Program, error) {
	if condition.OrgID.Valid {
		orgProperty, err := GetOrganizationPropertyModel().GetOrDefault(ctx, condition.OrgID.String)
		if err != nil {
			log.Error(ctx, "Query:GetOrganizationPropertyModel.GetOrDefault error", log.Err(err), log.Any("condition", condition))
			return nil, err
		}
		condition.OrgType = sql.NullString{
			String: string(orgProperty.Type),
			Valid:  true,
		}
	}

	var result []*entity.Program
	err := da.GetProgramDA().Query(ctx, condition, &result)
	if err != nil {
		log.Error(ctx, "query error", log.Err(err), log.Any("condition", condition))
		return nil, err
	}
	return result, nil
}

func (m *programModel) GetByID(ctx context.Context, id string) (*entity.Program, error) {
	var result = new(entity.Program)
	err := da.GetProgramDA().Get(ctx, id, result)
	if err == dbo.ErrRecordNotFound {
		log.Error(ctx, "GetByID:not found", log.Err(err), log.String("id", id))
		return nil, constant.ErrRecordNotFound
	}
	if err != nil {
		log.Error(ctx, "GetByID error", log.Err(err), log.String("id", id))
		return nil, err
	}
	return result, nil
}

var (
	_programOnce  sync.Once
	_programModel IProgramModel
)

func GetProgramModel() IProgramModel {
	_programOnce.Do(func() {
		_programModel = &programModel{}
	})
	return _programModel
}
