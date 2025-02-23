package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/VikaPaz/algalar/internal/models"
	"github.com/VikaPaz/algalar/internal/server/rest"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/tealeg/xlsx"
)

type Service interface {
	RegisterAuto(ctx context.Context, car models.Car) (models.Car, error)
	RegisterUser(ctx context.Context, user models.User) error
	UpdateUser(ctx context.Context, user models.User) (string, error)
	UpdateUserPassword(ctx context.Context, newPassword string) error
	GetUserDetails(ctx context.Context) (models.User, error)
	RegisterWheel(ctx context.Context, wheel models.Wheel) (models.Wheel, error)
	UpdateWheelData(ctx context.Context, wheel models.Wheel) error
	GetWheelData(ctx context.Context, id string) (models.Wheel, error)
	GenerateReport(ctx context.Context) ([]models.ReportData, error)
	IsCreatred(table string, key string, val any) (bool, error)
	GetAutoData(ctx context.Context, id string) (models.Car, error)
	GetAutoWheelsData(ctx context.Context, id string) (models.CarWithWheels, error)
	GetAutoList(ctx context.Context, offset int, limit int) ([]models.Car, error)
	GetWheelsData(ctx context.Context, stateNumber string) ([]models.Wheel, error)
	NewSensorData(ctx context.Context, newData models.SensorData) (models.SensorData, error)
	SensorsDataByCarID(ctx context.Context, carID string) ([]models.SensorsData, error)
	Temperaturedata(ctx context.Context, filter models.TemperatureDataByWheelIDFilter) ([]models.TemperatureData, error)
	Pressuredata(ctx context.Context, filter models.PressureDataByWheelIDFilter) ([]models.PressureData, error)
	CreateDriver(ctx context.Context, driver models.Driver) (models.Driver, error)
	GetAutoDataByStateNumber(ctx context.Context, stateNumber string) (models.Car, error)
	GetDriversList(ctx context.Context, offset int, limit int) ([]models.DriverStatisticsResponse, error)
	GetDriverInfo(ctx context.Context, driverID string) (models.DriverInfoResponse, error)
	GetDriverByCaDviceNum(ctx context.Context, deviceNum string) (models.Driver, error)
	UpdateDriverWorktime(ctx context.Context, deviceNum string, workedTime int) error
	CreatePosition(ctx context.Context, position models.Position) (models.Position, error)
	GetCarRoutePositions(ctx context.Context, carID string, from time.Time, to time.Time) ([]models.Position, error)
	GetCurrentCarPositions(ctx context.Context) ([]models.CurrentPositionResponse, error)
	GetCurrentCarPositionsByPoints(ctx context.Context, pointA models.Point, pointB models.Point) ([]models.CurrentPositionResponse, error)
	RegisterBeakege(ctx context.Context, breakege models.Breakage) (models.Breakage, error)
	CreateBreakageFromMqtt(ctx context.Context, breakage models.BreakageFromMqtt) (models.Breakage, error)
	GetBreakagesByCarId(ctx context.Context, carID string) ([]models.BreakageInfo, error)
	CreateNotification(ctx context.Context, new models.Notification) (models.Notification, error)
	UpdateNotificationStatus(ctx context.Context, id string, status string) error
	UpdateAllNotificationsStatus(ctx context.Context, status string) error
	GetNotificationInfo(ctx context.Context, notificationID string) (models.NotificationInfo, error)
	GetNotificationList(ctx context.Context, status string, limit, offset int) ([]models.NotificationListItem, error)
}

type AuthService interface {
	GenerateAccessToken(userID string) (string, error)
	GenerateRefreshToken(userID string) (string, *time.Time, error)
	ValidateRefreshToken(refreshToken string) (string, error)
	GetUserID(login, password string) (string, error)
	SaveRefreshToken(userID string, refreshToken string, expiration time.Time) error
	GetRefreshToken(userID string) (string, error)
	UpdateRefreshToken(userID string, token string, expiration time.Time) error
}

type ServImplemented struct {
	service Service
	auth    AuthService
	conf    Config
	log     *logrus.Logger
}

type Config struct {
	SigningKey string
}

func NewServer(conf Config, svc Service, auth AuthService, logger *logrus.Logger) *ServImplemented {
	return &ServImplemented{
		service: svc,
		auth:    auth,
		log:     logger,
		conf:    conf,
	}
}

// Auth
func (s *ServImplemented) PostLogin(w http.ResponseWriter, r *http.Request) {
	var loginDetails rest.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginDetails); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := s.auth.GetUserID(string(loginDetails.Email), loginDetails.Password)
	if err != nil {
		s.log.Error(err)
		if err == models.ErrNoContent {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	accessToken, err := s.auth.GenerateAccessToken(userID)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	refreshToken, exp, err := s.auth.GenerateRefreshToken(userID)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := s.auth.GetRefreshToken(userID)
	if err == models.ErrNoContent {
		err = nil
	}
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if token == "" {
		err = s.auth.SaveRefreshToken(userID, refreshToken, *exp)
	} else {
		err = s.auth.UpdateRefreshToken(userID, refreshToken, *exp)
	}
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := rest.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	w.WriteHeader(http.StatusCreated)
}

func (s *ServImplemented) PostRefresh(w http.ResponseWriter, r *http.Request) {
	headAuth, err := getTokenFromHeader(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	userID, err := s.auth.ValidateRefreshToken(headAuth)
	if err != nil {
		if err == models.ErrInvalidRefreshToken {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := s.auth.GetRefreshToken(userID)
	if err != nil {
		if err == models.ErrNoContent {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if token != headAuth {
		err := errors.New(fmt.Sprintf("refresh token not found or expired: %v", models.ErrInvalidRefreshToken))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	accessToken, err := s.auth.GenerateAccessToken(userID)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	refreshToken, exp, err := s.auth.GenerateRefreshToken(userID)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = s.auth.UpdateRefreshToken(userID, refreshToken, *exp)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := rest.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	w.WriteHeader(http.StatusCreated)
}

// User
func (s *ServImplemented) PostUser(w http.ResponseWriter, r *http.Request) {
	var userInfo rest.UserRegistration
	if err := json.NewDecoder(r.Body).Decode(&userInfo); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user models.User = ToNewUser(userInfo)

	ok, err := s.service.IsCreatred("users", "login", user.Login)
	if ok {
		err := models.ErrAlreadyExists
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.service.RegisterUser(r.Context(), user); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// Update user details
// (PUT /userinfo)
func (s *ServImplemented) PutUserinfo(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var userInfo rest.UserDetails
	if err := json.NewDecoder(r.Body).Decode(&userInfo); err != nil {
		s.log.Error(err)
		http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	var user models.User = models.User{
		INN:      *userInfo.Inn,
		Name:     *userInfo.FirstName,
		Surname:  *userInfo.LastName,
		Gender:   *userInfo.Gender,
		Login:    *userInfo.Email,
		Timezone: *userInfo.TimeZone,
		Phone:    *userInfo.Phone,
	}

	_, err = s.service.UpdateUser(ctx, user)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *ServImplemented) PutUser(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req rest.UpdatePassword
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.service.UpdateUserPassword(ctx, req.NewPassword); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *ServImplemented) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	user, err := s.service.GetUserDetails(ctx)
	if err != nil {
		s.log.Error(err)
		if err == models.ErrNoContent {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userDetails := ToUserDetails(user)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userDetails)
}

// Auto
func (s *ServImplemented) PostAuto(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req rest.AutoRegistration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user_id := ctx.Value("user_id").(string)
	car := ToCar(user_id, req)

	ok, err := s.service.IsCreatred("cars", "state_number", car.StateNumber)
	if ok {
		err := models.ErrAlreadyExists
		s.log.Error(err)
		http.Error(w, "already exists", http.StatusBadRequest)
		return
	}
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	car, err = s.service.RegisterAuto(ctx, car)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := ToAutoResponse(car)
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *ServImplemented) GetAuto(w http.ResponseWriter, r *http.Request, params rest.GetAutoParams) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	autoData, err := s.service.GetAutoData(ctx, params.CarId)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := ToAutoResponse(autoData)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *ServImplemented) GetAutoInfo(w http.ResponseWriter, r *http.Request, params rest.GetAutoInfoParams) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	autoWheelsData, err := s.service.GetAutoWheelsData(ctx, params.CarId)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	autoData := rest.AutoResponse{
		AxleCount:    &autoWheelsData.CountAxis,
		Brand:        &autoWheelsData.Brand,
		DeviceNumber: &autoWheelsData.DeviceNumber,
		Id:           &autoWheelsData.ID,
		StateNumber:  &autoWheelsData.StateNumber,
		UniqueId:     &autoWheelsData.IDUnicum,
		AutoType:     &autoWheelsData.AutoType,
	}
	resp := make(map[string]any)
	resp["auto"] = autoData
	countWheels := len(autoWheelsData.Wheels)
	wheels := make([]rest.WheelResponse, countWheels)
	for i := 0; i < countWheels; i++ {
		wheels[i] = ToWheelResponse(autoWheelsData.Wheels[i])
	}
	resp["wheels"] = wheels
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *ServImplemented) GetAutoList(w http.ResponseWriter, r *http.Request, params rest.GetAutoListParams) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	autoList, err := s.service.GetAutoList(ctx, params.Offset, params.Limit)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := make([]rest.AutoResponse, len(autoList))
	for i, val := range autoList {
		res[i] = ToAutoResponse(val)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Update car mileage
// (PUT /mileage)
func (s *ServImplemented) PutMileage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Wheel
func (s *ServImplemented) PostWheels(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req rest.WheelRegistration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var wheel models.Wheel = ToNewWheel(req)
	new, err := s.service.RegisterWheel(ctx, wheel)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := ToWheelResponse(new)

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *ServImplemented) PutWheels(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req rest.WheelChange
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var wheel models.Wheel = ToWheel(req)
	if err := s.service.UpdateWheelData(ctx, wheel); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var res rest.WheelResponse = ToWheelResponse(wheel)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *ServImplemented) GetWheels(w http.ResponseWriter, r *http.Request, params rest.GetWheelsParams) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	wheelData, err := s.service.GetWheelData(ctx, params.Id)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := ToWheelResponse(wheelData)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *ServImplemented) GetWheelsStateNumber(w http.ResponseWriter, r *http.Request, stateNumber string) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	dataList, err := s.service.GetWheelsData(ctx, stateNumber)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := make([]rest.WheelsDataForDevice, len(dataList))
	for i, val := range dataList {
		res[i] = ToWheelData(val)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// TODO:
// Sensor
// Update an existing sensor
// (POST /sensordata)
func (s *ServImplemented) PostSensordata(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req rest.NewSensorData
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var newData models.SensorData = ToNewData(req)

	_, err = s.service.NewSensorData(ctx, newData)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// Provides actual data by car ID
// (GET /sensors)
func (s *ServImplemented) GetSensors(w http.ResponseWriter, r *http.Request, params rest.GetSensorsParams) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	sensors, err := s.service.SensorsDataByCarID(ctx, params.CarId)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := make([]rest.SensorsData, len(sensors))
	for i, val := range sensors {
		res[i] = ToRestSensorsData(val)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// TODO:
// Data
// Get data by wheel ID
// (GET /temperaturedata)
func (s *ServImplemented) GetTemperaturedata(w http.ResponseWriter, r *http.Request, params rest.GetTemperaturedataParams) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	filter := models.TemperatureDataByWheelIDFilter{
		IDWheel: params.WheelId,
		From:    params.From,
		To:      params.To,
	}

	data, err := s.service.Temperaturedata(ctx, filter)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := make([]rest.TemperatureData, len(data))
	for i, val := range data {
		res[i] = rest.TemperatureData{
			Temperature: &val.Temperature,
			Time:        &val.Datetime,
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Get data by wheel ID
// (GET /pressuredata)
func (s *ServImplemented) GetPressuredata(w http.ResponseWriter, r *http.Request, params rest.GetPressuredataParams) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	filter := models.PressureDataByWheelIDFilter{
		IDWheel: params.WheelId,
		From:    params.From,
		To:      params.To,
	}

	data, err := s.service.Pressuredata(ctx, filter)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := make([]rest.PressureData, len(data))
	for i, val := range data {
		res[i] = rest.PressureData{
			Pressure: &val.Pressure,
			Time:     &val.Datetime,
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Driver
// Add a driver
// (POST /driver)
func (s *ServImplemented) PostDriver(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getUserID(r)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req rest.DriverRegistration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user_id := ctx.Value("user_id").(string)
	var newDriver models.Driver = ToNewDriver(user_id, req)

	car, err := s.service.GetAutoDataByStateNumber(ctx, req.StateNumber)
	if err != nil {
		s.log.Error(err)
		if err == models.ErrNoContent {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newDriver.IDCar = car.ID

	_, err = s.service.CreateDriver(ctx, newDriver)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Driver statistics
// (GET /driver/list)
func (s *ServImplemented) GetDriverList(w http.ResponseWriter, r *http.Request, params rest.GetDriverListParams) {
	ctx, err := s.getUserID(r)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	drivers, err := s.service.GetDriversList(ctx, params.Limit, params.Offset)
	if err != nil {
		s.log.Error(err)
		if err == models.ErrNoContent {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := make([]rest.DriverStatisticsResponse, len(drivers))

	for i, driver := range drivers {
		res[i] = ToDriverResponse(driver)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Driver information
// (GET /driver/info)
func (s *ServImplemented) GetDriverInfo(w http.ResponseWriter, r *http.Request, params rest.GetDriverInfoParams) {
	ctx, err := s.getUserID(r)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	driverID := params.DriverId

	driverInfo, err := s.service.GetDriverInfo(ctx, driverID.String())
	if err != nil {
		s.log.Error(err)
		if err == models.ErrDriverNotFound {
			http.Error(w, "Driver not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := rest.DriverInfoResponse{
		Name:       driverInfo.Name,
		Surname:    driverInfo.Surname,
		MiddleName: driverInfo.MiddleName,
		Phone:      driverInfo.Phone,
		Birthday:   driverInfo.Birthday,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.log.Error(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Update the driver's worked hours
// (PUT /driver/worktime)
func (s *ServImplemented) PutDriverWorktime(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getUserID(r)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var request rest.WorkTimeUpdateRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.log.Error(err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.DeviceNum == "" || request.WorkedTime <= 0 {
		s.log.Error("Missing or invalid fields")
		http.Error(w, "Missing or invalid fields", http.StatusBadRequest)
		return
	}

	err = s.service.UpdateDriverWorktime(ctx, request.DeviceNum, request.WorkedTime)
	if err != nil {
		s.log.Error(err)
		if err == models.ErrDriverNotFound {
			http.Error(w, "Driver not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Position
// Add car position from MQTT
// (POST /position)
func (s *ServImplemented) PostPosition(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getUserID(r)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req rest.PositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Error(err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Point) != 2 {
		err := fmt.Errorf("invalid point coordinates")
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	position := models.Position{
		DeviceNumber: req.DeviceNumber,
		Location: models.Point{
			Latitude:  float32(req.Point[0]),
			Longitude: float32(req.Point[1]),
		},
		CreatedAt: req.CreatedAt,
	}

	if _, err := s.service.CreatePosition(ctx, position); err != nil {
		s.log.Error(err)
		http.Error(w, "failed to add car position"+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// Get the route of a car
// (GET /position/carroute)
func (s *ServImplemented) GetPositionCarroute(w http.ResponseWriter, r *http.Request, params rest.GetPositionCarrouteParams) {
	ctx, err := s.getUserID(r)
	if err != nil {
		s.log.Errorf("%v: %v", models.ErrUnauthorized, err)
		http.Error(w, models.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	s.log.Debugf("Received request to fetch car route for car_id=%s, time_from=%v, time_to=%v", params.CarId, params.TimeFrom, params.TimeTo)

	carInfo, err := s.service.GetAutoData(ctx, params.CarId.String())
	if err != nil {
		s.log.Errorf("%v: %v", models.ErrFailedToFetchCarData, err)
		http.Error(w, models.ErrFailedToFetchCarData.Error(), http.StatusInternalServerError)
		return
	}

	s.log.Debugf("Fetched car info: %+v", carInfo)

	var res = rest.RouteCarResponse{
		Brand:       carInfo.Brand,
		StateNumber: carInfo.StateNumber,
		UniqueId:    carInfo.IDUnicum,
	}

	res.CarId, err = uuid.Parse(carInfo.ID)
	if err != nil {
		s.log.Errorf("%v: %v", models.ErrInvalidCarID, err)
		http.Error(w, models.ErrInvalidCarID.Error(), http.StatusInternalServerError)
		return
	}

	s.log.Debugf("Fetching route positions for car_id=%s", params.CarId)

	positions, err := s.service.GetCarRoutePositions(ctx, params.CarId.String(), params.TimeFrom, params.TimeTo)
	if err != nil {
		s.log.Errorf("%v: %v", models.ErrFailedToFetchRoutePositions, err)
		http.Error(w, models.ErrFailedToFetchRoutePositions.Error(), http.StatusInternalServerError)
		return
	}

	if len(positions) == 0 {
		s.log.Debugf("%v: No route positions found for car_id=%s", models.ErrNoContent, params.CarId)
		http.Error(w, models.ErrNoContent.Error(), http.StatusNoContent)
		return
	}

	s.log.Debugf("Successfully fetched %d route positions for car_id=%s", len(positions), params.CarId)

	route := make([]rest.Position, len(positions))
	for i, val := range positions {
		route[i] = rest.Position{
			Point:     []float32{val.Location.Latitude, val.Location.Longitude},
			CreatedAt: val.CreatedAt,
		}
	}

	res.Positions = route

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.log.Errorf("%v: %v", models.ErrFailedToEncodeResponse, err)
		http.Error(w, models.ErrFailedToEncodeResponse.Error(), http.StatusInternalServerError)
	}
}

// Get current car positions
// (GET /position/listcurrent)
func (s *ServImplemented) GetPositionListcurrent(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getUserID(r)
	if err != nil {
		s.log.Errorf("%v: %v", models.ErrUnauthorized, err)
		http.Error(w, models.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	s.log.Debugf("Received request to fetch current car positions for user_id=%s", ctx.Value("user_id"))

	positions, err := s.service.GetCurrentCarPositions(ctx)
	if err != nil {
		s.log.Errorf("%v: %v", models.ErrFailedToFetchPositions, err)
		http.Error(w, fmt.Sprintf("%v: %v", models.ErrFailedToFetchPositions, err), http.StatusInternalServerError)
		return
	}

	if len(positions) == 0 {
		s.log.Debugf("%v: No positions found for user_id=%s", models.ErrNoContent, ctx.Value("user_id"))
		http.Error(w, models.ErrNoContent.Error(), http.StatusNoContent)
		return
	}

	res := make([]rest.PositionCurrentListResponse, len(positions))
	for i, val := range positions {
		carID, err := uuid.Parse(val.IDCar)
		if err != nil {
			s.log.Errorf("%v: %v", models.ErrInvalidUUID, err)
			http.Error(w, fmt.Sprintf("%v: %v", models.ErrFailedToFetchPositions, err), http.StatusInternalServerError)
			return
		}
		res[i] = rest.PositionCurrentListResponse{
			CarId:    carID,
			Point:    []float32{val.Point.Latitude, val.Point.Longitude},
			UniqueId: val.IDUnicum,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.log.Errorf("%v: %v", models.ErrFailedToEncodeResponse, err)
		http.Error(w, fmt.Sprintf("%v: %v", models.ErrFailedToEncodeResponse, err), http.StatusInternalServerError)
	}
}

// Get list of cars
// (GET /positions/listcars)
func (s *ServImplemented) GetPositionsListcars(w http.ResponseWriter, r *http.Request, params rest.GetPositionsListcarsParams) {
	ctx, err := s.getUserID(r)
	if err != nil {
		s.log.Errorf("%v: %v", models.ErrUnauthorized, err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	s.log.Debugf("Fetching car positions list: userID=%s, offset=%d, limit=%d", ctx.Value("user_id"), params.Offset, params.Limit)

	cars, err := s.service.GetAutoList(ctx, params.Offset, params.Limit)
	if err != nil {
		s.log.Errorf("%v: %v", models.ErrFailedToFetchCars, err)
		http.Error(w, models.ErrFailedToFetchCars.Error(), http.StatusInternalServerError)
		return
	}

	if len(cars) == 0 {
		s.log.Debugf("No cars found for userID=%s", ctx.Value("user_id"))
		http.Error(w, models.ErrNoContent.Error(), http.StatusNoContent)
		return
	}

	s.log.Debugf("Successfully fetched %d cars for userID=%s", len(cars), ctx.Value("user_id"))

	res := make([]rest.PositionCarListResponse, len(cars))
	for i, val := range cars {
		id, err := uuid.Parse(val.ID)
		if err != nil {
			s.log.Errorf("%v: %v", models.ErrInvalidUUID, err)
			http.Error(w, models.ErrInvalidUUID.Error(), http.StatusInternalServerError)
			return
		}

		res[i] = rest.PositionCarListResponse{
			CarId:       &id,
			Brand:       &val.Brand,
			StateNumber: &val.StateNumber,
			UniqueId:    &val.IDUnicum,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.log.Errorf("%v: %v", models.ErrFailedToEncodeResponse, err)
		http.Error(w, models.ErrFailedToEncodeResponse.Error(), http.StatusInternalServerError)
	}
}

// Notifications
// Change the status of all notifications for a specific user
// (PUT /notification/allstatus)
func (s *ServImplemented) PutNotificationAllstatus(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getUserID(r)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req rest.ChangeAllNotificationsStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Error(err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		err := fmt.Errorf("missing required field: status")
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.service.UpdateAllNotificationsStatus(ctx, req.Status)
	if err != nil {
		s.log.Error(err)
		http.Error(w, "failed to update notifications status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Get detailed information about a specific notification
// (GET /notification/info)
func (s *ServImplemented) GetNotificationInfo(w http.ResponseWriter, r *http.Request, params rest.GetNotificationInfoParams) {
	ctx, err := s.getUserID(r)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if params.Id == uuid.Nil {
		err := fmt.Errorf("missing required parameter: id")
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var notificationInfo models.NotificationInfo

	notificationInfo, err = s.service.GetNotificationInfo(ctx, params.Id.String())
	if err != nil {
		if errors.Is(err, models.ErrNoContent) {
			s.log.Error(err)
			http.Error(w, "notification not found", http.StatusNotFound)
			return
		}
		s.log.Error(err)
		http.Error(w, "failed to fetch notification info", http.StatusInternalServerError)
		return
	}

	var res rest.NotificationInfoResponse = rest.NotificationInfoResponse{
		Description: notificationInfo.Description,
		DriverName:  notificationInfo.DriverName,
		Location: []float32{
			notificationInfo.Location.Latitude,
			notificationInfo.Location.Longitude,
		},
		CreatedAt: notificationInfo.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.log.Error(err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Get list of notifications based on status
// (GET /notification/list)
func (s *ServImplemented) GetNotificationList(w http.ResponseWriter, r *http.Request, params rest.GetNotificationListParams) {
	ctx, err := s.getUserID(r)
	if err != nil {
		s.log.Errorf("%v: %v", models.ErrUnauthorizedRequest, err)
		http.Error(w, models.ErrUnauthorizedRequest.Error(), http.StatusUnauthorized)
		return
	}

	s.log.Debugf("Received request to fetch notifications with status: %s, limit: %d, offset: %d", *params.Status, params.Limit, params.Offset)

	notifications, err := s.service.GetNotificationList(ctx, *params.Status, params.Limit, params.Offset)
	if err != nil {
		s.log.Errorf("%v: %v", models.ErrFailedToRetrieveNotifications, err)
		http.Error(w, models.ErrFailedToRetrieveNotifications.Error(), http.StatusInternalServerError)
		return
	}

	if len(notifications) == 0 {
		s.log.Debugf("%v: No notifications found", models.ErrNoContent)
		http.Error(w, models.ErrNoContent.Error(), http.StatusNoContent)
		return
	}

	s.log.Debugf("Successfully retrieved %d notifications", len(notifications))

	res := make([]rest.NotificationListResponse, len(notifications))
	for i, val := range notifications {
		res[i] = ToNotificationListResponse(val)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.log.Errorf("%v: %v", models.ErrFailedToEncodeResponse, err)
		http.Error(w, models.ErrFailedToEncodeResponse.Error(), http.StatusInternalServerError)
		return
	}

	s.log.Debugf("Response successfully encoded and sent")
	w.WriteHeader(http.StatusOK)
}

// Change the status of a specific notification
// (PUT /notification/status)
func (s *ServImplemented) PutNotificationStatus(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getUserID(r)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req rest.ChangeNotificationStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Error(err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Id == uuid.Nil || req.Status == "" {
		err := fmt.Errorf("missing required fields: id or status")
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.service.UpdateNotificationStatus(ctx, req.Id.String(), req.Status)
	if err != nil {
		if errors.Is(err, models.ErrNoContent) {
			s.log.Error(err)
			http.Error(w, "notification not found", http.StatusNotFound)
			return
		}
		s.log.Error(err)
		http.Error(w, "failed to update notification status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Breakages
// Add a new breakage from MQTT data
// (POST /breakage)
func (s *ServImplemented) PostBreakage(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getUserID(r)
	if err != nil {
		s.log.Errorf("%v: %v", models.ErrUnauthorizedRequest, err)
		http.Error(w, models.ErrUnauthorizedRequest.Error(), http.StatusUnauthorized)
		return
	}

	s.log.Debugf("Received request to create a breakage from user_id=%s", ctx.Value("user_id"))

	var req rest.BreakageFromMqttRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Errorf("%v: %v", models.ErrInvalidRequestBody, err)
		http.Error(w, models.ErrInvalidRequestBody.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Point) != 2 {
		s.log.Errorf("%v", models.ErrInvalidPointFormat)
		http.Error(w, models.ErrInvalidPointFormat.Error(), http.StatusBadRequest)
		return
	}

	point := models.Point{
		Latitude:  req.Point[0],
		Longitude: req.Point[1],
	}

	s.log.Debugf("Fetching driver by device number: %s", req.DeviceNum)
	driver, err := s.service.GetDriverByCaDviceNum(ctx, req.DeviceNum)
	if err != nil {
		s.log.Errorf("%v: %v", models.ErrFailedToFetchDriver, err)
		http.Error(w, models.ErrFailedToFetchDriver.Error(), http.StatusInternalServerError)
		return
	}

	breakage := models.Breakage{
		CarID:       driver.IDCar,
		DriverID:    driver.ID,
		Location:    point,
		Type:        req.Type,
		Description: req.Description,
		CreatedAt:   req.Datetime,
	}

	s.log.Debugf("Creating breakage: %+v", breakage)
	newBreakage, err := s.service.RegisterBeakege(ctx, breakage)
	if err != nil {
		s.log.Errorf("%v: %v", models.ErrFailedToCreateBreakage, err)
		http.Error(w, models.ErrFailedToCreateBreakage.Error(), http.StatusInternalServerError)
		return
	}

	notification := models.Notification{
		IDUser:     ctx.Value("user_id").(string),
		IDBreakage: newBreakage.ID,
		Note:       models.NoteBreakage + ": " + breakage.Description,
		Status:     models.StatusNew,
		CreatedAt:  time.Now(),
	}

	s.log.Debugf("Creating notification: %+v", notification)
	if _, err := s.service.CreateNotification(ctx, notification); err != nil {
		s.log.Errorf("%v: %v", models.ErrFailedToCreateNotification, err)
		http.Error(w, models.ErrFailedToCreateNotification.Error(), http.StatusInternalServerError)
		return
	}

	s.log.Debugf("Breakage and notification successfully created for user_id=%s", ctx.Value("user_id"))
	w.WriteHeader(http.StatusCreated)
}

// Get a list of breakages for a specific car
// (GET /breakage/list)
func (s *ServImplemented) GetBreakageList(w http.ResponseWriter, r *http.Request, params rest.GetBreakageListParams) {
	ctx, err := s.getUserID(r)
	if err != nil {
		s.log.Errorf("Unauthorized request: %v", err)
		http.Error(w, models.ErrUnauthorizedRequest.Error(), http.StatusUnauthorized)
		return
	}

	s.log.Debugf("Fetching breakages for car ID: %s", params.CarId.String())

	breakages, err := s.service.GetBreakagesByCarId(ctx, params.CarId.String())
	if err != nil {
		s.log.Errorf("Failed to fetch breakages: %v", err)
		http.Error(w, models.ErrFailedToFetchBreakages.Error(), http.StatusInternalServerError)
		return
	}

	res := make([]rest.BreakageListResponse, len(breakages))

	for i, val := range breakages {
		id, err := uuid.Parse(val.ID)
		if err != nil {
			s.log.Errorf("Failed to fetch breakages: %v", err)
			http.Error(w, models.ErrFailedToFetchBreakages.Error(), http.StatusInternalServerError)
			return
		}
		res[i] = rest.BreakageListResponse{
			Id:          &id,
			DriverName:  &val.DriverName,
			StateNumber: &val.StateNumber,
			Type:        &val.Type,
			Description: &val.Description,
			Datetime:    &val.CreatedAt,
		}
	}

	s.log.Debugf("Successfully fetched %d breakages for car ID: %s", len(breakages), params.CarId.String())

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.log.Errorf("Failed to encode response: %v", err)
		http.Error(w, models.ErrFailedToEncodeResponse.Error(), http.StatusInternalServerError)
	}
}

// Report
func (s *ServImplemented) GetReport(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	reportData, err := s.service.GenerateReport(ctx)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Report")
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	header := sheet.AddRow()
	header.AddCell().Value = "Id Wheel"
	header.AddCell().Value = "State Number"
	header.AddCell().Value = "Tire Brand"
	header.AddCell().Value = "Mileage"
	header.AddCell().Value = "Temp Out of Bounds"
	header.AddCell().Value = "Pressure Out of Bounds"

	for _, data := range reportData {
		row := sheet.AddRow()
		row.AddCell().Value = data.IdWheel
		row.AddCell().Value = data.StateNumber
		row.AddCell().Value = data.TireBrand
		row.AddCell().Value = strconv.FormatFloat(float64(data.Mileage), 'f', 2, 32)
		row.AddCell().Value = strconv.Itoa(data.TempOutOfBounds)
		row.AddCell().Value = strconv.Itoa(data.PressureOutOfBounds)
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=report.xlsx")

	err = file.Write(w)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// User
func ToNewUser(userRegistration rest.UserRegistration) models.User {
	return models.User{
		INN:      userRegistration.Inn,
		Name:     userRegistration.FirstName,
		Surname:  userRegistration.LastName,
		Gender:   userRegistration.Gender,
		Login:    userRegistration.Email,
		Password: userRegistration.Password,
		Timezone: userRegistration.TimeZone,
		Phone:    userRegistration.Phone,
	}
}

func ToUserDetails(user models.User) rest.UserDetails {
	return rest.UserDetails{
		Email:     &user.Login,
		FirstName: &user.Name,
		Inn:       &user.INN,
		LastName:  &user.Surname,
		Gender:    &user.Gender,
		Password:  &user.Password,
		Phone:     &user.Phone,
		TimeZone:  &user.Timezone,
	}
}

// Car
func ToCar(idCompany string, AutoRegistration rest.AutoRegistration) models.Car {
	return models.Car{
		IDCompany:    idCompany,
		StateNumber:  AutoRegistration.StateNumber,
		Brand:        AutoRegistration.Brand,
		DeviceNumber: AutoRegistration.DeviceNumber,
		IDUnicum:     AutoRegistration.UniqueId,
		CountAxis:    AutoRegistration.AxleCount,
		Type:         AutoRegistration.AutoType,
	}
}

func ToAutoResponse(car models.Car) rest.AutoResponse {
	return rest.AutoResponse{
		AxleCount:    &car.CountAxis,
		Brand:        &car.Brand,
		DeviceNumber: &car.DeviceNumber,
		Id:           &car.ID,
		StateNumber:  &car.StateNumber,
		UniqueId:     &car.IDUnicum,
		AutoType:     &car.Type,
	}
}

// Wheel
func ToNewWheel(new rest.WheelRegistration) models.Wheel {
	return models.Wheel{
		IDCar:          new.AutoId,
		AxisNumber:     new.AxleNumber,
		Position:       new.WheelPosition,
		SensorNumber:   new.SensorNumber,
		Size:           new.TireSize,
		Cost:           new.TireCost,
		Brand:          new.TireBrand,
		Model:          new.TireModel,
		Mileage:        new.Mileage,
		MinTemperature: new.MinTemperature,
		MinPressure:    new.MinPressure,
		MaxTemperature: new.MaxTemperature,
		MaxPressure:    new.MaxPressure,
		Ngp:            &new.Ngp,
		Tkvh:           &new.Tkvh,
	}
}

func ToWheel(wheel rest.WheelChange) models.Wheel {
	return models.Wheel{
		ID:             wheel.Id,
		IDCar:          wheel.AutoId,
		AxisNumber:     wheel.AxleNumber,
		Position:       wheel.WheelPosition,
		SensorNumber:   wheel.SensorNumber,
		Size:           wheel.TireSize,
		Cost:           wheel.TireCost,
		Brand:          wheel.TireBrand,
		Model:          wheel.TireModel,
		Mileage:        wheel.Mileage,
		MinTemperature: wheel.MinTemperature,
		MinPressure:    wheel.MinPressure,
		MaxTemperature: wheel.MaxTemperature,
		MaxPressure:    wheel.MaxPressure,
		Ngp:            &wheel.Ngp,
		Tkvh:           &wheel.Tkvh,
	}
}

func ToWheelResponse(wheel models.Wheel) rest.WheelResponse {
	return rest.WheelResponse{
		AxleNumber:     &wheel.AxisNumber,
		Id:             &wheel.ID,
		MaxPressure:    &wheel.MaxPressure,
		MaxTemperature: &wheel.MaxTemperature,
		Mileage:        &wheel.Mileage,
		MinPressure:    &wheel.MinPressure,
		MinTemperature: &wheel.MinTemperature,
		SensorNumber:   &wheel.SensorNumber,
		TireBrand:      &wheel.Brand,
		TireCost:       &wheel.Cost,
		TireModel:      &wheel.Model,
		TireSize:       &wheel.Size,
		AutoId:         &wheel.IDCar,
		WheelPosition:  &wheel.Position,
		Ngp:            wheel.Ngp,
		Tkvh:           wheel.Tkvh,
	}
}

func ToWheelData(wheel models.Wheel) rest.WheelsDataForDevice {
	return rest.WheelsDataForDevice{
		MaxPressure:    &wheel.MaxPressure,
		MaxTemperature: &wheel.MaxTemperature,
		MinPressure:    &wheel.MinPressure,
		MinTemperature: &wheel.MinTemperature,
		SensorNumber:   &wheel.SensorNumber,
		WheelPosition:  &wheel.Position,
	}
}

// Sensor
func ToRestSensorsData(data models.SensorsData) rest.SensorsData {
	return rest.SensorsData{
		WheelPosition: &data.WheelPosition,
		Pressure:      &data.Pressure,
		Temperature:   &data.Temperature,
	}
}

// Data
func ToNewData(data rest.NewSensorData) models.SensorData {
	return models.SensorData{
		SensorNumber: *data.SensorNumber,
		DeviceNumber: *data.DeviceNumber,
		Pressure:     *data.Pressure,
		Temperature:  *data.Temperature,
		Time:         *data.Time,
	}
}

// Driver
func ToNewDriver(idCompany string, new rest.DriverRegistration) models.Driver {
	return models.Driver{
		IDCompany: idCompany,
		Name:      new.Name,
		Surname:   new.Surname,
		Middle:    new.MiddleName,
		Phone:     new.Phone,
		Birthday:  new.Birthday.Time,
		Rating:    10,
	}
}

func ToDriverResponse(driver models.DriverStatisticsResponse) rest.DriverStatisticsResponse {
	return rest.DriverStatisticsResponse{
		BreakagesCount: driver.BreakagesCount,
		FullName:       driver.FullName,
		DriverId:       uuid.MustParse(driver.DriverID),
		Experience:     driver.Experience,
		Rating:         driver.Rating,
		WorkedTime:     driver.WorkedTime,
	}
}

func ToNotificationListResponse(new models.NotificationListItem) rest.NotificationListResponse {
	return rest.NotificationListResponse{
		Id:           uuid.MustParse(new.ID),
		StateNumber:  new.StateNumber,
		Brand:        new.Brand,
		BreakageType: new.BreakageType,
		CreatedAt:    new.CreatedAt,
	}
}

func validateToken(tokenStr string, jwtSecret string) (*models.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	if claims, ok := token.Claims.(*models.Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid claims")
}

func getTokenFromHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header missing")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}

func (s *ServImplemented) getUserID(r *http.Request) (context.Context, error) {
	tokenStr, err := getTokenFromHeader(r)
	if err != nil {
		return nil, fmt.Errorf("authorization: %w", err)
	}

	claims, err := validateToken(tokenStr, s.conf.SigningKey)
	if err != nil {
		return nil, fmt.Errorf("authorization: %w", err)
	}

	ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
	return ctx, nil
}
