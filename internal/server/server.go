package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/VikaPaz/algalar/internal/models"
	"github.com/VikaPaz/algalar/internal/server/rest"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
)

type Service interface {
	RegisterAuto(ctx context.Context, car models.Car) (models.Car, error)
	UserLogin(login string, password string) (string, string, error)
	RefreshToken(ctx context.Context) (string, string, error)
	RegisterUser(user models.User) error
	UpdateUserPassword(ctx context.Context, newPassword string) error
	GetUserDetails(ctx context.Context) (models.User, error)
	RegisterWheel(ctx context.Context, wheel models.Wheel) (models.Wheel, error)
	UpdateWheelData(ctx context.Context, wheel models.Wheel) error
	GetWheelData(ctx context.Context, id string) (models.Wheel, error)
	GenerateReport(ctx context.Context, params models.GetReportParams) (interface{}, error)
	GetSensorData(ctx context.Context, params models.GetSensorParams) (interface{}, error)
	IsCreatred(table string, key string, val any) (bool, error)
	GetAutoData(ctx context.Context, id string) (models.Car, error)
	GetAutoList(ctx context.Context, offset int, limit int) ([]models.Car, error)
	RegisterSensor(ctx context.Context, sensor models.Sensor) (models.Sensor, error)
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
		http.Error(w, "already exists", http.StatusBadRequest)
		return
	}
	if err != nil {
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
	ctx, err := getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	newAccess, errRefresh, err := s.service.RefreshToken(ctx)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	response := rest.TokenResponse{
		AccessToken:  newAccess,
		RefreshToken: errRefresh,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *ServImplemented) PostAuto(w http.ResponseWriter, r *http.Request) {
	ctx, err := getUserID(r)
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
		http.Error(w, "already exists", http.StatusBadRequest)
		return
	}
	if err != nil {
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
	ctx, err := getUserID(r)
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

func (s *ServImplemented) GetAutoList(w http.ResponseWriter, r *http.Request, params rest.GetAutoListParams) {
	ctx, err := getUserID(r)
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

// TODO:

// Register a new sensor
// (POST /sensor)
func (s *ServImplemented) PostSensor(w http.ResponseWriter, r *http.Request) {
	// ctx, err := getUserID(r)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusUnauthorized)
	// 	return
	// }

	// var Auto rest.SensorRegistration
	// if err := json.NewDecoder(r.Body).Decode(&Auto); err != nil {
	// 	s.log.Error(err)
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }

	// autoList, err := s.service.RegisterSensor(ctx)
	// if err != nil {
	// 	s.log.Error(err)
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// res := make([]rest.AutoResponse, len(autoList))
	// for i, val := range autoList {
	// 	res[i] = ToAutoResponse(val)
	// }

	// w.WriteHeader(http.StatusOK)
	// w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(res)

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

func (s *ServImplemented) PutUser(w http.ResponseWriter, r *http.Request) {
	ctx, err := getUserID(r)
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

func (s *ServImplemented) GetUser(w http.ResponseWriter, r *http.Request, params rest.GetUserParams) {
	ctx, err := getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	user, err := s.service.GetUserDetails(ctx)
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userDetails := ToUserDetails(user)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userDetails)
}

func (s *ServImplemented) PostWheels(w http.ResponseWriter, r *http.Request) {
	ctx, err := getUserID(r)
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
	ctx, err := getUserID(r)
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
	ctx, err := getUserID(r)
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

func (s *ServImplemented) GetReport(w http.ResponseWriter, r *http.Request, params rest.GetReportParams) {
	report, err := s.service.GenerateReport(r.Context(), models.GetReportParams{})
	if err != nil {
		s.log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(report.(string)))
}

func (s *ServImplemented) GetSensor(w http.ResponseWriter, r *http.Request, params rest.GetSensorParams) {
	data, err := s.service.GetSensorData(r.Context(), models.GetSensorParams{})
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

type tokenClaims struct {
	jwt.StandardClaims
	Data string `json:"data"`
}

var jwtSecret = []byte("qrkjk#4#%35FSFJlja#4353KSFjH")

func validateToken(tokenStr string) (*tokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	if claims, ok := token.Claims.(*tokenClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("invalid claims")
	}
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

func getUserID(r *http.Request) (context.Context, error) {
	tokenStr, err := getTokenFromHeader(r)
	if err != nil {
		return nil, fmt.Errorf("authorization: %w", err)
	}

	claims, err := validateToken(tokenStr)
	if err != nil {
		return nil, fmt.Errorf("authorization: %w", err)
	}

	ctx := context.WithValue(r.Context(), "user_id", claims.Data)
	return ctx, nil
}
