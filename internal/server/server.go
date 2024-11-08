package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/VikaPaz/algalar/internal/models"
	"github.com/VikaPaz/algalar/internal/server/rest"
	"github.com/sirupsen/logrus"
)

type Service interface {
	RegisterAuto(car models.Car) error
	UserLogin(login string, password string) (string, string, error)
	RefreshToken(token string) (string, string, error)
	RegisterUser(user models.User) error
	UpdateUserPassword(login string, newPassword string) error
	GetUserDetails(id string) (interface{}, error)
	RegisterWheel(wheel models.Wheel) error
	UpdateWheelData(wheel models.Wheel) error
	GetWheelData(id string) (interface{}, error)
	GenerateReport(params models.GetReportParams) (interface{}, error)
	GetSensorData(params models.GetSensorParams) (interface{}, error)
}

type ServImplemented struct {
	service Service
	log     *logrus.Logger
}

func NewServer(svc Service, logger *logrus.Logger) *ServImplemented {
	return &ServImplemented{
		service: svc,
		log:     logger,
	}
}

func (s *ServImplemented) PostAuto(w http.ResponseWriter, r *http.Request) {
	var Auto rest.AutoRegistration
	if err := json.NewDecoder(r.Body).Decode(&Auto); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	car := ToCar(Auto)
	if err := s.service.RegisterAuto(car); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *ServImplemented) GetAuto(w http.ResponseWriter, r *http.Request, params rest.GetAutoParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *ServImplemented) GetAutoList(w http.ResponseWriter, r *http.Request, params rest.GetAutoListParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Register a new sensor
// (POST /sensor)
func (s *ServImplemented) PostSensor(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Update an existing sensor
// (PUT /sensor)
func (s *ServImplemented) PutSensor(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Get breakages by car ID
// (GET /brackeges)
func (s *ServImplemented) GetBrackeges(w http.ResponseWriter, r *http.Request, params rest.GetBrackegesParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Register a new breakage
// (POST /brackeges)
func (s *ServImplemented) PostBrackeges(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *ServImplemented) PostLogin(w http.ResponseWriter, r *http.Request) {
	var loginDetails rest.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginDetails); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	access, refresh, err := s.service.UserLogin(string(loginDetails.Email), loginDetails.Password)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	response := rest.TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *ServImplemented) PostRefresh(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	var token string

	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	}
	fmt.Println(token)

	newAccess, errRefresh, err := s.service.RefreshToken(token)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	response := rest.TokenResponse{
		AccessToken:  newAccess,
		RefreshToken: errRefresh,
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *ServImplemented) PostUser(w http.ResponseWriter, r *http.Request) {
	var userInfo rest.UserRegistration
	if err := json.NewDecoder(r.Body).Decode(&userInfo); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.log.Debug(userInfo)

	user := ToUserRegistration(userInfo)
	if err := s.service.RegisterUser(user); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *ServImplemented) PutUser(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	var req rest.UpdatePassword
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.service.UpdateUserPassword(token, req.NewPassword); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *ServImplemented) GetUser(w http.ResponseWriter, r *http.Request, params rest.GetUserParams) {
	userDetails, err := s.service.GetUserDetails(params.Id)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userDetails)
}

func (s *ServImplemented) PostWheels(w http.ResponseWriter, r *http.Request) {
	var req rest.WheelRegistration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	wheel := ToWheel(req)

	if err := s.service.RegisterWheel(wheel); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *ServImplemented) PutWheels(w http.ResponseWriter, r *http.Request) {
	var req rest.WheelRegistration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	wheel := ToWheel(req)
	if err := s.service.UpdateWheelData(wheel); err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wheel)
}

func (s *ServImplemented) GetWheels(w http.ResponseWriter, r *http.Request, params rest.GetWheelsParams) {
	wheelData, err := s.service.GetWheelData(params.Id)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(wheelData.(string)))
}

func (s *ServImplemented) GetReport(w http.ResponseWriter, r *http.Request, params rest.GetReportParams) {
	report, err := s.service.GenerateReport(models.GetReportParams{})
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(report.(string)))
}

func (s *ServImplemented) GetSensor(w http.ResponseWriter, r *http.Request, params rest.GetSensorParams) {
	data, err := s.service.GetSensorData(models.GetSensorParams{})
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(data.(string)))
}

func ToUser(userDetails rest.UserDetails) models.User {
	var inn int
	if userDetails.Inn != nil {
		inn = 123456789
	}
	return models.User{
		ID:         "",
		INN:        inn,
		Name:       *userDetails.FirstName,
		Surname:    *userDetails.LastName,
		MiddleName: *userDetails.MiddleName,
		Login:      *userDetails.Email,
		Password:   *userDetails.Password,
		Timezone:   *userDetails.TimeZone,
	}
}

func ToUserDetails(user models.User) rest.UserDetails {
	return rest.UserDetails{
		Email:      &user.Login,
		FirstName:  &user.Name,
		Inn:        nil,
		LastName:   &user.Surname,
		MiddleName: &user.MiddleName,
		Password:   &user.Password,
		Phone:      nil,
		TimeZone:   &user.Timezone,
	}
}

func ToUserRegistration(userRegistration rest.UserRegistration) models.User {
	return models.User{
		ID:         "",
		INN:        123456789,
		Name:       userRegistration.FirstName,
		Surname:    userRegistration.LastName,
		MiddleName: *userRegistration.MiddleName,
		Login:      userRegistration.Email,
		Password:   userRegistration.Password,
		Timezone:   *userRegistration.TimeZone,
	}
}

func ToUserRegistrationFromUser(user models.User) rest.UserRegistration {
	return rest.UserRegistration{
		Email:      user.Login,
		FirstName:  user.Name,
		Inn:        nil,
		LastName:   user.Surname,
		MiddleName: &user.MiddleName,
		Password:   user.Password,
		Phone:      "",
		TimeZone:   &user.Timezone,
	}
}

func ToCar(AutoRegistration rest.AutoRegistration) models.Car {
	return models.Car{
		ID:          AutoRegistration.UniqueId,
		IDCompany:   AutoRegistration.CompanyInn,
		StateNumber: AutoRegistration.StateNumber,
		Brand:       AutoRegistration.Brand,
		IDDevice:    AutoRegistration.DeviceId,
		IDUnicum:    AutoRegistration.UniqueId,
		CountAxis:   AutoRegistration.AxleCount,
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
		AutoType:    nil,
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
		TireSize:       nil,
		AutoId:         &wheel.IDCar,
		WheelPosition:  &wheel.Position,
	}
}
