package intergrate_academic_profile

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/constant"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

func (s *MapperImpl) Subject(ctx context.Context, organizationID, programID, subjectID string) (string, error) {
	amsSubjectID, found := s.subjectMapping[programID+":"+subjectID]
	if found {
		return amsSubjectID, nil
	}

	s.subjectMutex.Lock()
	defer s.subjectMutex.Unlock()

	ourSubject, found := s.ourSubjects[subjectID]
	if !found {
		log.Error(ctx, "unknown subject id", log.String("subjectID", subjectID), log.String("organizationID", organizationID))
		return "", constant.ErrRecordNotFound
	}

	amsProgramID, err := s.Program(ctx, organizationID, programID)
	if err != nil {
		return "", err
	}

	amsSubject, found := s.amsSubjects[amsProgramID+":"+ourSubject.Name]
	if found {
		s.subjectMapping[programID+":"+ourSubject.ID] = amsSubject.ID
		return amsSubject.ID, nil
	}

	// new subject id should be mapping
	subjects, err := external.GetSubjectServiceProvider().GetByProgram(ctx, s.operator, amsProgramID)
	if err != nil {
		return "", err
	}

	s.amsSubjects = make(map[string]*external.Subject, len(subjects))
	for _, subject := range subjects {
		s.amsSubjects[amsProgramID+":"+subject.Name] = subject
	}

	amsSubject, found = s.amsSubjects[amsProgramID+":"+ourSubject.Name]
	if !found {
		log.Error(ctx, "unknown subject id",
			log.String("organizationID", organizationID),
			log.String("programID", programID),
			log.Any("subject", ourSubject))

		return "", constant.ErrRecordNotFound
	}

	s.subjectMapping[programID+":"+ourSubject.ID] = amsSubject.ID
	return amsSubject.ID, nil
}

func (s *MapperImpl) initSubjectMapper(ctx context.Context) error {
	s.subjectMapping = make(map[string]string)
	s.amsSubjects = make(map[string]*external.Subject)

	return s.loadOurSubjects(ctx)
}

func (s *MapperImpl) loadOurSubjects(ctx context.Context) error {
	s.ourSubjects = map[string]*entity.Subject{
		"subject0": {
			ID:   "subject0",
			Name: "None Specified",
			// Name: "Non specified",
		},
		"subject1": {
			ID:   "subject1",
			Name: "Language/Literacy",
		},
		"subject2": {
			ID:   "subject2",
			Name: "Math",
		},
		"subject3": {
			ID:   "subject3",
			Name: "Science",
		},
	}
	// var subjects []*entity.Subject
	// err := da.GetSubjectDA().Query(ctx, &da.SubjectCondition{}, &subjects)
	// if err != nil {
	// 	log.Error(ctx, "query error", log.Err(err))
	// 	return err
	// }

	// // key: subject name
	// s.ourSubjects = make(map[string]*entity.Subject, len(subjects))
	// for _, subject := range subjects {
	// 	s.ourSubjects[subject.Name] = subject
	// }

	return nil
}
