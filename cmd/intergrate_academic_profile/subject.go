package intergrate_academic_profile

import (
	"context"

	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/entity"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/external"
)

func (s *MapperImpl) Subject(ctx context.Context, organizationID, programID, subjectID string) (string, error) {
	s.subjectMutex.Lock()
	defer s.subjectMutex.Unlock()

	amsSubjectID, found := s.subjectMapping[programID+":"+subjectID]
	if found {
		return amsSubjectID, nil
	}

	ourSubject, found := s.ourSubjects[subjectID]
	if !found {
		log.Error(ctx, "unknown subject id", log.String("subjectID", subjectID), log.String("organizationID", organizationID))
		return s.defaultSubject(programID)
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

		return s.defaultSubject(programID)
	}

	s.subjectMapping[programID+":"+ourSubject.ID] = amsSubject.ID
	return amsSubject.ID, nil
}

func (s *MapperImpl) initSubjectMapper(ctx context.Context) error {
	log.Info(ctx, "init prgram cache start")
	defer log.Info(ctx, "init program cache end")

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

func (s MapperImpl) defaultSubject(programID string) (string, error) {
	switch programID {
	case "5fd9ddface9660cbc5f667d8":
		return "5e9a201e-9c2f-4a92-bb6f-1ccf8177bb71", nil
	case "5fdac06ea878718a554ff00d":
		return "20d6ca2f-13df-4a7a-8dcb-955908db7baa", nil
	case "5fdac0f61f066722a1351adb":
		return "7cf8d3a3-5493-46c9-93eb-12f220d101d0", nil
	case "5fdac0fe1f066722a1351ade":
		return "fab745e8-9e31-4d0c-b780-c40120c98b27", nil
	case "program0":
		return "5e9a201e-9c2f-4a92-bb6f-1ccf8177bb71", nil
	case "program1":
		return "f037ee92-212c-4592-a171-ed32fb892162", nil
	case "program2":
		return "36c4f793-9aa3-4fb8-84f0-68a2ab920d5a", nil
	case "program3":
		return "29d24801-0089-4b8e-85d3-77688e961efb", nil
	case "program4":
		return "66a453b0-d38f-472e-b055-7a94a94d66c4", nil
	case "program5":
		return "b997e0d1-2dd7-40d8-847a-b8670247e96b", nil
	case "program6":
		return "b19f511e-a46b-488d-9212-22c0369c8afd", nil
	case "program7":
		return "49c8d5ee-472b-47a6-8c57-58daf863c2e1", nil
	default:
		return "5e9a201e-9c2f-4a92-bb6f-1ccf8177bb71", nil
	}
}
