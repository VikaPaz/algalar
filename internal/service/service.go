package service

import (
	"context"
	"fmt"
	"time"

	"github.com/VikaPaz/algalar/internal/models"
	"github.com/sirupsen/logrus"
)

type Repository interface {
	CreateUser(user models.User) (string, error)
	UpdateUser(user models.User) (string, error)
	GetById(userID string) (models.User, error)
	ChangePassword(userID, newPassword string) error
	GetIDByLoginAndPassword(login, password string) (string, error)
	CreateCar(car models.Car) (models.Car, error)
	CreateWheel(wheel models.Wheel) (string, error)
	GetWheelById(wheelID string) (models.Wheel, error)
	ChangeWheel(wheel models.Wheel) error
	SelectAny(table string, key string, val any) (bool, error)
	CreateBreakage(breakage models.Breakage) (string, error)
	GetCarById(carID string) (models.Car, error)
	GetCarByStateNumber(stateNumber string) (models.Car, error)
	GetCarsList(user_id string, offset int, limit int) ([]models.Car, error)
	GetIdCarByStateNumber(stateNumber string) (string, error)
	GetBreakagesByCarId(carID string) ([]models.BreakageInfo, error)
	GetReportData(userId string) ([]models.ReportData, error)
	GetWheelsByStateNumber(stateNumber string) ([]models.Wheel, error)
	GetCarWheelData(carID string) (models.CarWithWheels, error)
	CreateData(newData models.SensorData) (models.SensorData, error)
	SensorsDataByCarID(carID string) ([]models.SensorsData, error)
	Temperaturedata(filter models.TemperatureDataByWheelIDFilter) ([]models.TemperatureData, error)
	Pressuredata(filter models.PressureDataByWheelIDFilter) ([]models.PressureData, error)
	CreateDriver(driver models.Driver) (models.Driver, error)
	GetDriversList(user_id string, limit int, offset int) ([]models.DriverStatisticsResponse, error)
	GetDriverInfo(driverID string) (models.DriverInfoResponse, error)
	UpdateDriverWorktime(deviceNum string, workedTime int) error
	CreatePosition(ctx context.Context, position models.Position) (models.Position, error)
	GetCarRoutePositions(ctx context.Context, carID string, from time.Time, to time.Time) ([]models.Position, error)
	GetCurrentCarPositions(ctx context.Context, pointA models.Point, pointB models.Point) ([]models.CurentPosition, error)
	CreateBreakageFromMqtt(ctx context.Context, breakage models.BreakageFromMqtt) (models.Breakage, error)
	CreateNotification(new models.Notification) (models.Notification, error)
	UpdateNotificationStatus(ctx context.Context, id string, status string) error
	UpdateAllNotificationsStatus(ctx context.Context, userID string, status string) error
	GetNotificationInfo(ctx context.Context, notificationID string) (models.NotificationInfo, error)
	GetNotificationList(ctx context.Context, status string, limit, offset int) ([]models.NotificationListItem, error)
	CheckDriverExists(ctx context.Context, deviceNumber string) (bool, error)
}

type Service struct {
	repo Repository
	log  *logrus.Logger
}

func (s *Service) IsCreatred(table string, key string, val any) (bool, error) {
	ok, err := s.repo.SelectAny(table, key, val)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (s *Service) RegisterUser(ctx context.Context, user models.User) error {
	if user.Login == "" || user.Password == "" {
		return models.ErrLoginOrPassword
	}

	_, err := s.repo.CreateUser(user)
	if err != nil {
		s.log.Debugf("Error creating user: %s", user.Login)
		return err
	}
	return nil
}

// UpdateUser updates user information and returns the updated user ID.
func (s *Service) UpdateUser(ctx context.Context, user models.User) (string, error) {
	user_id, ok := ctx.Value("user_id").(string)
	if !ok {
		s.log.Errorf("Invalid context: %v", ctx)
		return "", fmt.Errorf("%w: %v", models.ErrInvalidContext, ctx)
	}
	user.ID = user_id

	s.log.Debugf("Updating user with ID: %s", user.ID)

	res, err := s.repo.UpdateUser(user)
	if err != nil {
		s.log.Errorf("Failed to update user: %v", err)
		return "", fmt.Errorf("%w: %v", models.ErrUserUpdateFailed, err)
	}

	s.log.Debugf("User updated successfully: %s", res)
	return res, nil
}

func (s *Service) RegisterAuto(ctx context.Context, car models.Car) (models.Car, error) {
	id, ok := ctx.Value("user_id").(string)
	if !ok {
		return models.Car{}, fmt.Errorf("wrong context: %v", ctx)
	}
	car.IDCompany = id

	res, err := s.repo.CreateCar(car)
	if err != nil {
		s.log.Debugf("Error registering Auto: %v", car)
		return models.Car{}, err
	}
	return res, nil
}

func (s *Service) UpdateUserPassword(ctx context.Context, newPassword string) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return fmt.Errorf("wrong context: %v", ctx)
	}

	err := s.repo.ChangePassword(userID, newPassword)
	if err != nil {
		s.log.Debugf("Error updating user password: %s", userID)
		return err
	}

	s.log.Debugf("User password updated successfully: %s", userID)
	return nil
}

func (s *Service) GetUserDetails(ctx context.Context) (models.User, error) {
	id, ok := ctx.Value("user_id").(string)
	if !ok {
		return models.User{}, fmt.Errorf("wrong context: %v", ctx)
	}
	user, err := s.repo.GetById(id)
	if err != nil {
		s.log.Debugf("User not found: %s", id)
		return models.User{}, err
	}

	s.log.Debugf("User details fetched successfully: %s", id)
	return user, nil
}

func (s *Service) RegisterWheel(ctx context.Context, wheel models.Wheel) (models.Wheel, error) {
	id, ok := ctx.Value("user_id").(string)
	if !ok {
		return models.Wheel{}, fmt.Errorf("wrong context: %v", ctx)
	}
	wheel.IDCompany = id
	id_wheel, err := s.repo.CreateWheel(wheel)
	if err != nil {
		s.log.Debugf("Error registering wheel: %v", wheel)
		return models.Wheel{}, err
	}
	wheel.ID = id_wheel

	s.log.Debugf("Wheel registered successfully: %v", wheel)
	return wheel, nil
}

func (s *Service) RegisterBeakege(ctx context.Context, breakege models.Breakage) (models.Breakage, error) {
	id, ok := ctx.Value("user_id").(string)
	if !ok {
		return models.Breakage{}, fmt.Errorf("wrong context: %v", ctx)
	}
	id_wheel, err := s.repo.CreateBreakage(breakege)
	if err != nil {
		s.log.Debugf("Error registering sensor: %v", id)
		return models.Breakage{}, err
	}
	breakege.ID = id_wheel

	s.log.Debugf("Sensor registered successfully: %v", id)
	return breakege, nil
}

func (s *Service) UpdateWheelData(ctx context.Context, wheel models.Wheel) error {
	err := s.repo.ChangeWheel(wheel)
	if err != nil {
		s.log.Debugf("Error updating wheel data: %v", wheel)
		return err
	}

	s.log.Debugf("Wheel data updated successfully: %v", wheel)
	return nil
}

func (s *Service) GetWheelData(ctx context.Context, id string) (models.Wheel, error) {
	wheel, err := s.repo.GetWheelById(id)
	if err != nil {
		s.log.Debugf("Wheel not found: %s", id)
		return models.Wheel{}, err
	}

	s.log.Debugf("Wheel data fetched successfully: %s", id)
	return wheel, nil
}

func (s *Service) GetWheelsData(ctx context.Context, stateNumber string) ([]models.Wheel, error) {
	data, err := s.repo.GetWheelsByStateNumber(stateNumber)
	if err != nil {
		s.log.Debugf("Auto not found: %s", stateNumber)
		return []models.Wheel{}, err
	}

	s.log.Debugf("Auto data fetched successfully: %s", stateNumber)
	return data, nil
}

func (s *Service) GetAutoData(ctx context.Context, id string) (models.Car, error) {
	auto, err := s.repo.GetCarById(id)
	if err != nil {
		s.log.Debugf("Auto not found: %s", id)
		return models.Car{}, err
	}

	s.log.Debugf("Auto data fetched successfully: %s", id)
	return auto, nil
}

func (s *Service) GetAutoDataByStateNumber(ctx context.Context, stateNumber string) (models.Car, error) {
	auto, err := s.repo.GetCarByStateNumber(stateNumber)
	if err != nil {
		s.log.Debugf("Auto not found: %s", stateNumber)
		return models.Car{}, err
	}

	s.log.Debugf("Auto data fetched successfully: %s", stateNumber)
	return auto, nil
}

func (s *Service) GetAutoList(ctx context.Context, offset int, limit int) ([]models.Car, error) {
	user_id, ok := ctx.Value("user_id").(string)
	if !ok {
		return []models.Car{}, fmt.Errorf("wrong context: %v", ctx)
	}

	list, err := s.repo.GetCarsList(user_id, offset, limit)
	if err != nil {
		s.log.Debugf("not found: %s", user_id)
		return []models.Car{}, err
	}

	s.log.Debugf("data fetched successfully: %s", user_id)
	return list, nil
}

func (s *Service) GetCarId(ctx context.Context, stateNumber string) (string, error) {
	id, err := s.repo.GetIdCarByStateNumber(stateNumber)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *Service) GenerateReport(ctx context.Context) ([]models.ReportData, error) {
	repost, err := s.repo.GetReportData(ctx.Value("user_id").(string))
	if err != nil {
		return []models.ReportData{}, err
	}
	return repost, nil
}

func (s *Service) GetBreakagesByCarId(ctx context.Context, carID string) ([]models.BreakageInfo, error) {
	list, err := s.repo.GetBreakagesByCarId(carID)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *Service) GetAutoWheelsData(ctx context.Context, id string) (models.CarWithWheels, error) {
	resp, err := s.repo.GetCarWheelData(id)
	if err != nil {
		return models.CarWithWheels{}, err
	}

	return resp, nil
}

func (s *Service) NewSensorData(ctx context.Context, newData models.SensorData) (models.SensorData, error) {
	res, err := s.repo.CreateData(newData)
	if err != nil {
		return models.SensorData{}, err
	}
	return res, nil
}

func (s *Service) SensorsDataByCarID(ctx context.Context, carID string) ([]models.SensorsData, error) {
	res, err := s.repo.SensorsDataByCarID(carID)
	if err != nil {
		return []models.SensorsData{}, err
	}
	return res, nil
}

func (s *Service) Temperaturedata(ctx context.Context, filter models.TemperatureDataByWheelIDFilter) ([]models.TemperatureData, error) {
	res, err := s.repo.Temperaturedata(filter)
	if err != nil {
		return []models.TemperatureData{}, err
	}
	return res, nil
}

func (s *Service) Pressuredata(ctx context.Context, filter models.PressureDataByWheelIDFilter) ([]models.PressureData, error) {
	res, err := s.repo.Pressuredata(filter)
	if err != nil {
		return []models.PressureData{}, err
	}
	return res, nil
}

// Driver
func (s *Service) CreateDriver(ctx context.Context, driver models.Driver) (models.Driver, error) {
	res, err := s.repo.CreateDriver(driver)
	if err != nil {
		return models.Driver{}, err
	}
	return res, nil
}

func (s *Service) GetDriversList(ctx context.Context, limit int, offset int) ([]models.DriverStatisticsResponse, error) {
	res, err := s.repo.GetDriversList(ctx.Value("user_id").(string), limit, offset)
	if err != nil {
		return []models.DriverStatisticsResponse{}, err
	}
	return res, nil
}

func (s *Service) GetDriverInfo(ctx context.Context, driverID string) (models.DriverInfoResponse, error) {
	res, err := s.repo.GetDriverInfo(driverID)
	if err != nil {
		return models.DriverInfoResponse{}, err
	}
	return res, nil
}

func (s *Service) UpdateDriverWorktime(ctx context.Context, deviceNum string, workedTime int) error {
	err := s.repo.UpdateDriverWorktime(deviceNum, workedTime)
	if err != nil {
		return err
	}
	return nil
}

// Position
func (s *Service) CreatePosition(ctx context.Context, position models.Position) (models.Position, error) {
	position, err := s.repo.CreatePosition(ctx, position)
	if err != nil {
		return models.Position{}, err
	}
	return position, nil
}

func (s *Service) GetCarRoutePositions(ctx context.Context, carID string, from time.Time, to time.Time) ([]models.Position, error) {
	positions, err := s.repo.GetCarRoutePositions(ctx, carID, from, to)
	if err != nil {
		return []models.Position{}, err
	}
	return positions, nil
}

func (s *Service) GetCurrentCarPositions(ctx context.Context, pointA models.Point, pointB models.Point) ([]models.CurentPosition, error) {
	positions, err := s.repo.GetCurrentCarPositions(ctx, pointA, pointB)
	if err != nil {
		return []models.CurentPosition{}, err
	}
	return positions, nil
}

func (s *Service) CreateBreakageFromMqtt(ctx context.Context, breakage models.BreakageFromMqtt) (models.Breakage, error) {
	s.log.Debugf("Creating breakage from MQTT: %+v", breakage)

	ok, err := s.repo.CheckDriverExists(ctx, breakage.DeviceNum)
	if err != nil {
		s.log.Errorf("%v: %v", models.ErrFailedToCreateBreakage, err)
		return models.Breakage{}, models.ErrFailedToCreateBreakage
	}
	if !ok {
		s.log.Errorf("%v: %v", models.ErrFailedToCreateBreakage, err)
		return models.Breakage{}, models.ErrFailedToCreateBreakage
	}

	res, err := s.repo.CreateBreakageFromMqtt(ctx, breakage)
	if err != nil {
		s.log.Errorf("%v: %v", models.ErrFailedToCreateBreakage, err)
		return models.Breakage{}, models.ErrFailedToCreateBreakage
	}

	s.log.Debugf("Breakage created successfully: %+v", res)
	return res, nil
}

func (s *Service) CreateNotification(ctx context.Context, new models.Notification) (models.Notification, error) {
	res, err := s.repo.CreateNotification(new)
	if err != nil {
		return models.Notification{}, err
	}
	return res, nil
}

func (s *Service) UpdateNotificationStatus(ctx context.Context, id string, status string) error {
	err := s.repo.UpdateNotificationStatus(ctx, id, status)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateAllNotificationsStatus(ctx context.Context, status string) error {
	id, ok := ctx.Value("user_id").(string)
	if !ok {
		return fmt.Errorf("wrong context: %v", ctx)
	}

	err := s.repo.UpdateAllNotificationsStatus(ctx, id, status)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetNotificationInfo(ctx context.Context, notificationID string) (models.NotificationInfo, error) {
	if notificationID == "" {
		return models.NotificationInfo{}, fmt.Errorf("notification ID is required")
	}

	notificationInfo, err := s.repo.GetNotificationInfo(ctx, notificationID)
	if err != nil {
		return models.NotificationInfo{}, fmt.Errorf("failed to retrieve notification info: %w", err)
	}

	return notificationInfo, nil
}

func (s *Service) GetNotificationList(ctx context.Context, status string, limit, offset int) ([]models.NotificationListItem, error) {
	if status == "" {
		return nil, fmt.Errorf("status is required")
	}

	notifications, err := s.repo.GetNotificationList(ctx, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve notifications: %w", err)
	}

	return notifications, nil
}

func NewService(repo Repository, log *logrus.Logger) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}
