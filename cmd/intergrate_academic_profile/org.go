package intergrate_academic_profile

func (s MapperImpl) IsHeaderQuarter(organizationID string) bool {
	return s.headquarters[organizationID]
}
