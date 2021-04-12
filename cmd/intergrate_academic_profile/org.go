package intergrate_academic_profile

func (s *MapperImpl) IsHeaderQuarter(organizationID string) bool {
	return s.headquarters[organizationID]
}

func (s *MapperImpl) OrganizationType(organizationID string) string {
	if s.IsHeaderQuarter(organizationID) {
		return "HQ"
	}

	return "Non-HQ"
}
