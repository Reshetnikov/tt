// For all go:build
// If a function is defined in a file without a build tag, but is used in a file with a build tag, it is considered unused. Therefore, functions defined here are public.
package dashboard

import (
	"os"
	"time"

	"github.com/stretchr/testify/mock"
)

func SetAppDir() {
	os.Chdir("/app")
}

type MockDashboardRepository struct {
	mock.Mock
}

func (m *MockDashboardRepository) Tasks(userID int, taskCompleted string) (tasks []*Task) {
	args := m.Called(userID, taskCompleted)
	return args.Get(0).([]*Task)
}

func (m *MockDashboardRepository) TaskByID(id int) *Task {
	args := m.Called(id)
	return args.Get(0).(*Task)
}

func (m *MockDashboardRepository) CreateTask(task *Task) (int, error) {
	args := m.Called(task)
	return args.Int(0), args.Error(1)
}

func (m *MockDashboardRepository) UpdateTask(task *Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockDashboardRepository) DeleteTask(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockDashboardRepository) GetMaxSortOrder(userId int, isCompleted bool) (maxSortOrder int) {
	args := m.Called(userId, isCompleted)
	return args.Int(0)
}

func (m *MockDashboardRepository) UpdateTaskSortOrder(taskID, userID, sortOrder int) error {
	args := m.Called(taskID, userID, sortOrder)
	return args.Error(0)
}

func (m *MockDashboardRepository) RecordsWithTasks(filterRecords FilterRecords) (records []*Record) {
	args := m.Called(filterRecords)
	return args.Get(0).([]*Record)
}

func (m *MockDashboardRepository) RecordByIDWithTask(recordID int) *Record {
	args := m.Called(recordID)
	return args.Get(0).(*Record)
}

func (m *MockDashboardRepository) CreateRecord(record *Record) (int, error) {
	args := m.Called(record)
	return args.Int(0), args.Error(1)
}

func (m *MockDashboardRepository) UpdateRecord(record *Record) error {
	args := m.Called(record)
	return args.Error(0)
}

func (m *MockDashboardRepository) DeleteRecord(recordID int) error {
	args := m.Called(recordID)
	return args.Error(0)
}

func (m *MockDashboardRepository) DailyRecords(filterRecords FilterRecords, nowWithTimezone time.Time) (dailyRecords []DailyRecords) {
	args := m.Called(filterRecords, nowWithTimezone)
	return args.Get(0).([]DailyRecords)
}

func (m *MockDashboardRepository) Reports(userID int, startInterval time.Time, endInterval time.Time, nowWithTimezone time.Time) ReportData {
	args := m.Called(userID, startInterval, endInterval, nowWithTimezone)
	return args.Get(0).(ReportData)
}
