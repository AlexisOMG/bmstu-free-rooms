package service

type Service struct {
	scheduleStorage ScheduleStorage
}

func NewService(scheduleStorage ScheduleStorage) *Service {
	return &Service{
		scheduleStorage: scheduleStorage,
	}
}
