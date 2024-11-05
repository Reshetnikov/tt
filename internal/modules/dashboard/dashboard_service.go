package dashboard

import (
	"time"
)

type DashboardService struct {
	repo *DashboardRepositoryPostgres
}

// NewDashboardService создает новый сервис для работы с данными на панели управления
func NewDashboardService(repo *DashboardRepositoryPostgres) *DashboardService {
	return &DashboardService{repo: repo}
}

// GetDashboardData извлекает данные для отображения на панели управления
func (s *DashboardService) GetDashboardData(userID int) (DashboardData, error) {
	tasks, err := s.repo.FetchTasks(userID)
	if err != nil {
		return DashboardData{}, err
	}

	selectedWeek := time.Now().Truncate(24*time.Hour).AddDate(0, 0, -int(time.Now().Weekday())) // Начало недели
	weeklyRecords, err := s.repo.FetchWeeklyRecords(userID, selectedWeek)
	if err != nil {
		return DashboardData{}, err
	}

	return DashboardData{
		Tasks:         tasks,
		WeeklyRecords: weeklyRecords,
		SelectedWeek:  selectedWeek,
	}, nil
}
