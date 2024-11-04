package models

type User struct {
	INN        int
	Name       string
	Surname    string
	MiddleName string
	Login      string
	Password   string
	Timezone   string
}

type Car struct {
	IDCompany   string
	StateNamber string
	Brand       string
	IDDevice    int
	IDUnicum    int
	CountAxis   int
}

type Wheel struct {
	IDCar          string
	AxisNumber     int
	Position       int
	Size           float64
	Cost           int
	Brand          string
	Model          string
	Mileage        int
	MinTemperature int
	MinPressure    float64
	MaxTemperature int
	MaxPressure    float64
}
