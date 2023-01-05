package services

type ServiceStatusCondition struct {
	Service        *Service
	ExpectedStatus ServiceStatus
}

func (c *ServiceStatusCondition) IsMet() bool {
	return c.ExpectedStatus == c.Service.Status
}
