package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/VikaPaz/algalar/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	repo := NewRepository(db, logger)

	user := models.User{
		INN:      "1234567890",
		Name:     "John",
		Surname:  "Doe",
		Gender:   "M",
		Login:    "johndoe",
		Password: "password",
		Timezone: 0, // replace 0 with the appropriate integer value for the timezone
		Phone:    "123456789",
	}

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(user.INN, user.Name, user.Surname, user.Gender, user.Login, user.Password, user.Timezone, user.Phone).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

	userID, err := repo.CreateUser(user)
	assert.NoError(t, err)
	assert.Equal(t, "1", userID)
}

func TestUpdateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	repo := NewRepository(db, logger)

	user := models.User{
		ID:       "1",
		INN:      "1234567890",
		Name:     "John",
		Surname:  "Doe",
		Gender:   "M",
		Login:    "johndoe",
		Timezone: 0, // replace 0 with the appropriate integer value for the timezone
		Phone:    "123456789",
	}

	mock.ExpectQuery("UPDATE users").
		WithArgs(user.INN, user.Name, user.Surname, user.Gender, user.Login, user.Timezone, user.Phone, user.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

	userID, err := repo.UpdateUser(user)
	assert.NoError(t, err)
	assert.Equal(t, "1", userID)
}

func TestGetById(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	repo := NewRepository(db, logger)

	userID := "1"
	expectedUser := models.User{
		INN:      "1234567890",
		Name:     "John",
		Surname:  "Doe",
		Gender:   "M",
		Login:    "johndoe",
		Password: "password",
		Timezone: 0, // replace 0 with the appropriate integer value for the timezone
		Phone:    "123456789",
	}

	mock.ExpectQuery("SELECT inn, name, surname, gender, login, password, utc_timezone, phone FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"inn", "name", "surname", "gender", "login", "password", "utc_timezone", "phone"}).
			AddRow(expectedUser.INN, expectedUser.Name, expectedUser.Surname, expectedUser.Gender, expectedUser.Login, expectedUser.Password, expectedUser.Timezone, expectedUser.Phone))

	user, err := repo.GetById(userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestChangePassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	repo := NewRepository(db, logger)

	userID := "1"
	newPassword := "newpassword"

	mock.ExpectExec("UPDATE users SET password = \\$1 WHERE id = \\$2").
		WithArgs(newPassword, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.ChangePassword(userID, newPassword)
	assert.NoError(t, err)
}

func TestGetIDByLoginAndPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	repo := NewRepository(db, logger)

	email := "johndoe"
	password := "password"
	expectedUserID := "1"

	mock.ExpectQuery("SELECT id FROM users WHERE login = \\$1 AND password = \\$2").
		WithArgs(email, password).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedUserID))

	userID, err := repo.GetIDByLoginAndPassword(email, password)
	assert.NoError(t, err)
	assert.Equal(t, expectedUserID, userID)
}

func TestCreateCar(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	repo := NewRepository(db, logger)

	car := models.Car{
		IDCompany:    "1",
		StateNumber:  "ABC123",
		Brand:        "Toyota",
		DeviceNumber: "12345",
		IDUnicum:     "unique123",
		CountAxis:    4,
		Type:         "SUV",
	}

	mock.ExpectQuery("INSERT INTO cars").
		WithArgs(car.IDCompany, car.StateNumber, car.Brand, car.DeviceNumber, car.IDUnicum, car.CountAxis, car.Type).
		WillReturnRows(sqlmock.NewRows([]string{"id", "id_company", "state_number", "brand", "device_number", "id_unicum", "car_type", "count_axis"}).
			AddRow("1", car.IDCompany, car.StateNumber, car.Brand, car.DeviceNumber, car.IDUnicum, car.Type, car.CountAxis))

	createdCar, err := repo.CreateCar(car)
	assert.NoError(t, err)
	assert.Equal(t, car.IDCompany, createdCar.IDCompany)
	assert.Equal(t, car.StateNumber, createdCar.StateNumber)
	assert.Equal(t, car.Brand, createdCar.Brand)
	assert.Equal(t, car.DeviceNumber, createdCar.DeviceNumber)
	assert.Equal(t, car.IDUnicum, createdCar.IDUnicum)
	assert.Equal(t, car.Type, createdCar.Type)
	assert.Equal(t, car.CountAxis, createdCar.CountAxis)
}

func TestGetCarById(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	repo := NewRepository(db, logger)

	carID := "1"
	expectedCar := models.Car{
		ID:           carID,
		IDCompany:    "1",
		StateNumber:  "ABC123",
		Brand:        "Toyota",
		DeviceNumber: "12345",
		IDUnicum:     "unique123",
		CountAxis:    4,
	}

	mock.ExpectQuery("SELECT id, id_company, state_number, brand, device_number, id_unicum, count_axis FROM cars WHERE id = \\$1").
		WithArgs(carID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "id_company", "state_number", "brand", "device_number", "id_unicum", "count_axis"}).
			AddRow(expectedCar.ID, expectedCar.IDCompany, expectedCar.StateNumber, expectedCar.Brand, expectedCar.DeviceNumber, expectedCar.IDUnicum, expectedCar.CountAxis))

	car, err := repo.GetCarById(carID)
	assert.NoError(t, err)
	assert.Equal(t, expectedCar, car)
}

func TestGetCarByDeviceNumber(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	repo := NewRepository(db, logger)

	deviceNumber := "12345"
	expectedCar := models.Car{
		ID:           "1",
		IDCompany:    "1",
		StateNumber:  "ABC123",
		Brand:        "Toyota",
		DeviceNumber: deviceNumber,
		IDUnicum:     "unique123",
		Type:         "SUV",
		CountAxis:    4,
	}

	mock.ExpectQuery("SELECT id, id_company, state_number, brand, device_number, id_unicum, car_type, count_axis FROM cars WHERE device_number = \\$1").
		WithArgs(deviceNumber).
		WillReturnRows(sqlmock.NewRows([]string{"id", "id_company", "state_number", "brand", "device_number", "id_unicum", "car_type", "count_axis"}).
			AddRow(expectedCar.ID, expectedCar.IDCompany, expectedCar.StateNumber, expectedCar.Brand, expectedCar.DeviceNumber, expectedCar.IDUnicum, expectedCar.Type, expectedCar.CountAxis))

	car, err := repo.GetCarByDeviceNumber(context.Background(), deviceNumber)
	assert.NoError(t, err)
	assert.Equal(t, expectedCar, car)
}

func TestGetIdCarByStateNumber(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	repo := NewRepository(db, logger)

	stateNumber := "ABC123"
	expectedCarID := "1"

	mock.ExpectQuery("SELECT id FROM cars WHERE state_number = \\$1").
		WithArgs(stateNumber).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedCarID))

	carID, err := repo.GetIdCarByStateNumber(stateNumber)
	assert.NoError(t, err)
	assert.Equal(t, expectedCarID, carID)
}

func TestGetCarsList(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	logger := logrus.New()
	repo := NewRepository(db, logger)

	userID := "1"
	offset := 0
	limit := 10
	expectedCars := []models.Car{
		{
			ID:           "1",
			IDCompany:    userID,
			StateNumber:  "ABC123",
			Brand:        "Toyota",
			DeviceNumber: "12345",
			IDUnicum:     "unique123",
			CountAxis:    4,
		},
	}

	mock.ExpectQuery("SELECT id, id_company, state_number, brand, device_number, id_unicum, count_axis FROM cars WHERE id_company = \\$1 LIMIT \\$2 OFFSET \\$3").
		WithArgs(userID, limit, offset).
		WillReturnRows(sqlmock.NewRows([]string{"id", "id_company", "state_number", "brand", "device_number", "id_unicum", "count_axis"}).
			AddRow(expectedCars[0].ID, expectedCars[0].IDCompany, expectedCars[0].StateNumber, expectedCars[0].Brand, expectedCars[0].DeviceNumber, expectedCars[0].IDUnicum, expectedCars[0].CountAxis))

	cars, err := repo.GetCarsList(userID, offset, limit)
	assert.NoError(t, err)
	assert.Equal(t, expectedCars, cars)
}
