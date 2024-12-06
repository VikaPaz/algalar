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
	"github.com/sirupsen/logrus"
	"github.com/tealeg/xlsx"
)

type Service interface {
	RegisterAuto(ctx context.Context, car models.Car) (models.Car, error)
	RegisterUser(user models.User) error
	UpdateUserPassword(ctx context.Context, newPassword string) error
	GetUserDetails(ctx context.Context) (models.User, error)
	RegisterWheel(ctx context.Context, wheel models.Wheel) (models.Wheel, error)
	UpdateWheelData(ctx context.Context, wheel models.Wheel) error
	GetWheelData(ctx context.Context, id string) (models.Wheel, error)
	GenerateReport(ctx context.Context, userId string) ([]models.ReportData, error)
	GetSensorData(ctx context.Context, id string) ([]models.Sensor, error)
	GetBreackegeData(ctx context.Context, id string) ([]models.Breakage, error)
	IsCreatred(table string, key string, val any) (bool, error)
	GetAutoData(ctx context.Context, id string) (models.Car, error)
	GetAutoWheelsData(ctx context.Context, id string) (models.CarWithWheels, error)
	GetAutoList(ctx context.Context, offset int, limit int) ([]models.Car, error)
	RegisterSensor(ctx context.Context, sensor models.Sensor) (models.Sensor, error)
	RegisterBeakege(ctx context.Context, breakege models.Breakage) (models.Breakage, error)
	GetCarId(stateNumber string) (string, error)
	UpdateSensor(ctx context.Context, sensor models.Sensor) (models.Sensor, error)
	GetWheelsData(ctx context.Context, stateNumber string) ([]models.Wheel, error)
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

func (s *ServImplemented) PostUser(w http.ResponseWriter, r *http.Request) {
	var userInfo rest.UserRegistration
	if err := json.NewDecoder(r.Body).Decode(&userInfo); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := ToUserRegistration(userInfo)

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

	if err := s.service.RegisterUser(user); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

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
}

func (s *ServImplemented) PostAuto(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var Auto rest.AutoRegistration
	if err := json.NewDecoder(r.Body).Decode(&Auto); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	car := ToCar(Auto)

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

	response := ToAutoResponse(car)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
		AxleCount:   &autoWheelsData.CountAxis,
		Brand:       &autoWheelsData.Brand,
		DeviceId:    &autoWheelsData.IDDevice,
		Id:          &autoWheelsData.ID,
		StateNumber: &autoWheelsData.StateNumber,
		UniqueId:    &autoWheelsData.IDUnicum,
		AutoType:    &autoWheelsData.AutoType,
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

	autoList, err := s.service.GetAutoList(ctx, *params.Offset, *params.Limit)
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

	wheel := ToWheel(req)
	new, err := s.service.RegisterWheel(ctx, wheel)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := ToWheelResponse(new)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *ServImplemented) PutWheels(w http.ResponseWriter, r *http.Request) {
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

	wheel := ToWheel(req)
	if err := s.service.UpdateWheelData(ctx, wheel); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wheel)
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
	dataList, err := s.service.GetWheelsData(r.Context(), stateNumber)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := make([]rest.WheelResponse, len(dataList))
	for i, val := range dataList {
		res[i] = ToWheelResponse(val)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *ServImplemented) GetSensor(w http.ResponseWriter, r *http.Request, params rest.GetSensorParams) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	dataList, err := s.service.GetSensorData(ctx, params.CarId)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := make([]rest.SensorData, len(dataList))
	for i, val := range dataList {
		res[i] = ToSensorData(val, val.Datetime.String())
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Register a new sensor
// (POST /sensor)
func (s *ServImplemented) PostSensor(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req rest.SensorRegistration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := s.service.GetCarId(*req.StateNumber)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sensor := ToSensor(req, id)

	new_sensor, err := s.service.RegisterSensor(ctx, sensor)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := ToSensorData(new_sensor, new_sensor.Datetime.String())
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Update an existing sensor
// (PUT /sensor)
func (s *ServImplemented) PutSensor(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req rest.SensorRegistration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := s.service.GetCarId(*req.StateNumber)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sensor := ToSensor(req, id)

	data, err := s.service.UpdateSensor(ctx, sensor)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := ToSensorData(data, data.Datetime.String())
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Get breakages by car ID
// (GET /brackeges)
func (s *ServImplemented) GetBreakages(w http.ResponseWriter, r *http.Request, params rest.GetBreakagesParams) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	dateList, err := s.service.GetBreackegeData(ctx, params.CarId)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := make([]rest.BreakageResponse, len(dateList))
	for i, val := range dateList {
		res[i] = ToBreakageResponse(val)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Register a new breakage
// (POST /brackeges)
func (s *ServImplemented) PostBreakages(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req rest.BreakageRegistration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := s.service.GetCarId(*req.StateNumber)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	breakege := ToBreakage(req, id)

	new_breackege, err := s.service.RegisterBeakege(ctx, breakege)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := ToBreakageResponse(new_breackege)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *ServImplemented) GetReport(w http.ResponseWriter, r *http.Request, params rest.GetReportParams) {
	ctx, err := s.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	reportData, err := s.service.GenerateReport(ctx, params.UserId)
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

func ToUser(userDetails rest.UserDetails) models.User {
	return models.User{
		ID:       "",
		INN:      *userDetails.Inn,
		Name:     *userDetails.FirstName,
		Surname:  *userDetails.LastName,
		Gender:   *userDetails.Gender,
		Login:    *userDetails.Email,
		Password: *userDetails.Password,
		Timezone: *userDetails.TimeZone,
		Phone:    *userDetails.Phone,
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

func ToUserRegistration(userRegistration rest.UserRegistration) models.User {
	return models.User{
		ID:       "",
		INN:      string(userRegistration.Inn),
		Name:     userRegistration.FirstName,
		Surname:  userRegistration.LastName,
		Gender:   userRegistration.Gender,
		Login:    userRegistration.Email,
		Password: userRegistration.Password,
		Timezone: userRegistration.TimeZone,
		Phone:    userRegistration.Phone,
	}
}

func ToUserRegistrationFromUser(user models.User) rest.UserRegistration {
	return rest.UserRegistration{
		Email:     user.Login,
		FirstName: user.Name,
		Inn:       string(user.INN),
		LastName:  user.Surname,
		Gender:    user.Gender,
		Password:  user.Password,
		Phone:     user.Phone,
		TimeZone:  user.Timezone,
	}
}

func ToCar(AutoRegistration rest.AutoRegistration) models.Car {
	return models.Car{
		IDCompany:   AutoRegistration.CompanyInn,
		StateNumber: AutoRegistration.StateNumber,
		Brand:       AutoRegistration.Brand,
		IDDevice:    AutoRegistration.DeviceId,
		IDUnicum:    AutoRegistration.UniqueId,
		CountAxis:   AutoRegistration.AxleCount,
		Type:        AutoRegistration.AutoType,
	}
}

func ToAutoResponse(car models.Car) rest.AutoResponse {
	return rest.AutoResponse{
		AxleCount:   &car.CountAxis,
		Brand:       &car.Brand,
		CompanyInn:  &car.IDCompany,
		DeviceId:    &car.IDDevice,
		Id:          &car.ID,
		StateNumber: &car.StateNumber,
		UniqueId:    &car.IDUnicum,
		AutoType:    &car.Type,
	}
}

func ToSensor(sensorReg rest.SensorRegistration, carID string) models.Sensor {
	return models.Sensor{
		CarID:       carID,
		StateNumber: *sensorReg.StateNumber,
		CountAxis:   *sensorReg.CountAxis,
		Position:    *sensorReg.Position,
		Pressure:    *sensorReg.Pressure,
		Temperature: *sensorReg.Temperature,
	}
}

func ToBreakage(breakageResponse rest.BreakageRegistration, id string) models.Breakage {
	return models.Breakage{
		ID:          "",
		CarID:       id,
		StateNumber: *breakageResponse.StateNumber,
		Type:        *breakageResponse.Type,
		Description: *breakageResponse.Description,
		Datetime:    *breakageResponse.Datetime,
	}
}

func ToWheel(wheelRegistration rest.WheelRegistration) models.Wheel {
	return models.Wheel{
		ID:             wheelRegistration.SensorNumber,
		IDCar:          wheelRegistration.AutoId,
		AxisNumber:     wheelRegistration.AxleNumber,
		Position:       wheelRegistration.WheelPosition,
		Size:           0,
		Cost:           wheelRegistration.TireCost,
		Brand:          wheelRegistration.TireBrand,
		Model:          wheelRegistration.TireModel,
		Mileage:        wheelRegistration.Mileage,
		MinTemperature: wheelRegistration.MinTemperature,
		MinPressure:    wheelRegistration.MinPressure,
		MaxTemperature: wheelRegistration.MaxTemperature,
		MaxPressure:    wheelRegistration.MaxPressure,
		Ngp:            &wheelRegistration.Ngp,
		Tkvh:           &wheelRegistration.Tkvh,
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
		SensorNumber:   &wheel.ID,
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

func ToSensorData(sensor models.Sensor, time string) rest.SensorData {
	return rest.SensorData{
		Position:    &sensor.Position,
		Pressure:    &sensor.Pressure,
		Temperature: &sensor.Temperature,
		Time:        &time,
	}
}

func ToBreakageResponse(breakage models.Breakage) rest.BreakageResponse {
	datetimeStr := ""
	if !breakage.Datetime.IsZero() {
		datetimeStr = breakage.Datetime.Format(time.RFC3339)
	}

	return rest.BreakageResponse{
		Id:          &breakage.ID,
		StateNumber: &breakage.StateNumber,
		Type:        &breakage.Type,
		Description: &breakage.Description,
		Datetime:    &datetimeStr,
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
