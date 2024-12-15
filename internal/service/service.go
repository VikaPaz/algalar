package service

import (
	"context"
	"fmt"

	"github.com/VikaPaz/algalar/internal/models"
	"github.com/sirupsen/logrus"
)

type Repository interface {
	CreateUser(user models.User) (string, error)
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
	GetCarsList(user_id string, offset int, limit int) ([]models.Car, error)
	GetIdCarByStateNumber(stateNumber string) (string, error)
	GetBreakagesByCarId(carID string) ([]models.Breakage, error)
	GetReportData(userId string) ([]models.ReportData, error)
	GetWheelsByStateNumber(stateNumber string) ([]models.Wheel, error)
	GetCarWheelData(carID string) (models.CarWithWheels, error)
	CreateData(newData models.SensorData) (models.SensorData, error)
	SensorsDataByCarID(carID string) ([]models.SensorsData, error)
	Temperaturedata(filter models.TemperatureDataByWheelIDFilter) ([]models.TemperatureData, error)
	Pressuredata(filter models.PressureDataByWheelIDFilter) ([]models.PressureData, error)
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
	_, err := s.repo.CreateUser(user)
	if err != nil {
		s.log.Debugf("Error creating user: %s", user.Login)
		return err
	}
	return nil
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

func (s *Service) GetBreackegeData(ctx context.Context, carID string) ([]models.Breakage, error) {
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
func NewService(repo Repository, log *logrus.Logger) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}
