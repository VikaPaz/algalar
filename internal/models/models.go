package models

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type User struct {
	ID       string
	INN      string
	Name     string
	Surname  string
	Gender   string
	Login    string
	Password string
	Timezone int
	Phone    string
}

type Car struct {
	ID           string
	IDCompany    string
	StateNumber  string
	Brand        string
	DeviceNumber string
	IDUnicum     string
	CountAxis    int
	Type         string
}

type Wheel struct {
	ID             string
	IDCompany      string
	IDCar          string
	AxisNumber     int
	SensorNumber   string
	Position       int
	Size           float32
	Cost           float32
	Brand          string
	Model          string
	Mileage        float32
	MinTemperature float32
	MinPressure    float32
	MaxTemperature float32
	MaxPressure    float32
	Ngp            *float32
	Tkvh           *float32
}

type SensorData struct {
	ID           string
	SensorNumber string
	DeviceNumber string
	Pressure     float32
	Temperature  float32
	Time         time.Time
}

type SensorsData struct {
	WheelPosition int
	Pressure      float32
	Temperature   float32
}

type TemperatureData struct {
	Temperature float32
	Datetime    time.Time
}

type PressureData struct {
	Pressure float32
	Datetime time.Time
}

type TemperatureDataByWheelIDFilter struct {
	IDWheel string
	From    time.Time
	To      time.Time
}

type PressureDataByWheelIDFilter struct {
	IDWheel string
	From    time.Time
	To      time.Time
}

type Driver struct {
	ID         string
	IDCompany  string
	IDCar      string
	Name       string
	Surname    string
	Middle     string
	Phone      string
	Birthday   time.Time
	Rating     float32
	WorkedTime int
	CreatedAt  time.Time
}

type DriverStatisticsResponse struct {
	FullName       string
	WorkedTime     int
	Experience     float32
	Rating         float32
	BreakagesCount int
	DriverID       string
}

type DriverInfoResponse struct {
	Name       string
	Surname    string
	MiddleName string
	Phone      string
	Birthday   time.Time
}

type Position struct {
	ID           string
	DeviceNumber string
	Location     Point
	CreatedAt    time.Time
}

type Point struct {
	Latitude  float32
	Longitude float32
}

type CurentPosition struct {
	Point    Point
	IDCar    string
	IDUnicum string
}

type Breakage struct {
	ID          string
	CarID       string
	DriverID    string
	Location    Point
	Type        string
	Description string
	CreatedAt   time.Time
}

type BreakageFromMqtt struct {
	DeviceNum   string
	Type        string
	Description string
	CreatedAt   time.Time
	Point       [2]float32 // (latitude, longitude)
}

type BreakageInfo struct {
	ID          string
	DriverName  string
	StateNumber string
	Type        string
	Description string
	CreatedAt   time.Time
}

type Notification struct {
	ID         string
	IDUser     string
	IDBreakage string
	Note       string
	Status     string
	CreatedAt  time.Time
}

type NotificationInfo struct {
	Description string    `json:"description"`
	DriverName  string    `json:"driver_name"`
	Location    Point     `json:"location"`
	CreatedAt   time.Time `json:"created_at"`
}

type NotificationListItem struct {
	ID           string    `json:"id"`
	StateNumber  string    `json:"state_number"`
	Brand        string    `json:"brand"`
	BreakageType string    `json:"breakage_type"`
	CreatedAt    time.Time `json:"created_at"`
}

type GetReportParams struct {
	UserId string `form:"userId" json:"userId"`
}

type GetSensorParams struct {
	WheelId string `form:"wheelId" json:"wheelId"`
}

type ReportData struct {
	IdWheel             string
	StateNumber         string
	TireBrand           string
	Mileage             float32
	TempOutOfBounds     int
	PressureOutOfBounds int
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type CarWithWheels struct {
	ID           string  `json:"id"`
	StateNumber  string  `json:"state_number"`
	Brand        string  `json:"brand"`
	DeviceNumber string  `json:"id_device"`
	IDUnicum     string  `json:"id_unicum"`
	CountAxis    int     `json:"count_axis"`
	AutoType     string  `json:"AutoType"`
	Wheels       []Wheel `json:"wheels"`
}
