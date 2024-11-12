package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/VikaPaz/algalar/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Repository struct {
	conn *sql.DB
	log  *logrus.Logger
}

func NewRepository(conn *sql.DB, logger *logrus.Logger) *Repository {
	return &Repository{
		conn: conn,
		log:  logger,
	}
}

func (r *Repository) CreateUser(user models.User) (string, error) {
	if user.Login == "" || user.Password == "" {
		return "", errors.New("login and password are required")
	}

	query := `
        INSERT INTO users (inn, name, surname, middle_name, login, password, timezone)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id`

	var userID string
	err := r.conn.QueryRow(query, user.INN, user.Name, user.Surname, user.MiddleName, user.Login, user.Password, user.Timezone).Scan(&userID)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func (r *Repository) GetById(userID string) (models.User, error) {
	query := `
        SELECT inn, name, surname, middle_name, login, password, timezone
        FROM users
        WHERE id = $1`

	user := models.User{}
	err := r.conn.QueryRow(query, userID).Scan(&user.INN, &user.Name, &user.Surname, &user.MiddleName, &user.Login, &user.Password, &user.Timezone)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, nil
		}
		return models.User{}, err
	}

	return user, nil
}

func (r *Repository) ChangePassword(userID, newPassword string) error {
	if newPassword == "" {
		return errors.New("new password is required")
	}

	query := `
        UPDATE users
        SET password = $1
        WHERE id = $2`

	_, err := r.conn.Exec(query, newPassword, userID)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetIDByLoginAndPassword(email, password string) (string, error) {
	if email == "" || password == "" {
		return "", errors.New("email and password are required")
	}

	query := `
        SELECT id
        FROM users
        WHERE login = $1 AND password = $2`

	var userID string
	err := r.conn.QueryRow(query, email, password).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return userID, nil
}

func (r *Repository) CreateCar(car models.Car) (models.Car, error) {
	query := `
        INSERT INTO cars (id_company, state_number, brand, id_device, id_unicum, count_axis)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING *`
	resp := models.Car{}
	err := r.conn.QueryRow(query, car.IDCompany, car.StateNumber, car.Brand, car.IDDevice, car.IDUnicum, car.CountAxis).Scan(
		&resp.ID,
		&resp.IDCompany,
		&resp.StateNumber,
		&resp.Brand,
		&resp.IDDevice,
		&resp.IDUnicum,
		&resp.CountAxis,
	)
	if err != nil {
		return models.Car{}, err
	}

	return car, nil
}

func (r *Repository) CreateWheel(wheel models.Wheel) (string, error) {
	query := `
        INSERT INTO wheels (id_company, id_car, count_axis, position, size, cost, brand, model, mileage, min_temperature, min_pressure, max_temperature, max_pressure)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
        RETURNING id`

	var wheelID string
	err := r.conn.QueryRow(query, wheel.IDCompany, wheel.IDCar, wheel.AxisNumber, wheel.Position, wheel.Size, wheel.Cost, wheel.Brand, wheel.Model, wheel.Mileage, wheel.MinTemperature, wheel.MinPressure, wheel.MaxTemperature, wheel.MaxPressure).Scan(&wheelID)
	if err != nil {
		return "", err
	}

	return wheelID, nil
}

func (r *Repository) GetWheelById(wheelID string) (models.Wheel, error) {
	query := `
        SELECT id_car, count_axis, position, size, cost, brand, model, mileage, min_temperature, min_pressure, max_temperature, max_pressure
        FROM wheels
        WHERE id = $1`

	wheel := models.Wheel{}
	err := r.conn.QueryRow(query, wheelID).Scan(&wheel.IDCar, &wheel.AxisNumber, &wheel.Position, &wheel.Size, &wheel.Cost, &wheel.Brand, &wheel.Model, &wheel.Mileage, &wheel.MinTemperature, &wheel.MinPressure, &wheel.MaxTemperature, &wheel.MaxPressure)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Wheel{}, nil
		}
		return models.Wheel{}, err
	}

	return wheel, nil
}

func (r *Repository) GetCarById(carID string) (models.Car, error) {
	query := `
        SELECT id, id_company, state_number, brand, id_device, id_unicum, count_axis
        FROM cars
        WHERE id = $1`

	car := models.Car{}
	err := r.conn.QueryRow(query, carID).Scan(
		&car.ID,
		&car.IDCompany,
		&car.StateNumber,
		&car.Brand,
		&car.IDDevice,
		&car.IDUnicum,
		&car.CountAxis,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.Car{}, nil
		}
		return models.Car{}, err
	}

	return car, nil
}

func (r *Repository) GetIdCarByStateNumber(stateNumber string) (string, error) {
	query := `
        SELECT id
        FROM cars
        WHERE state_number = $1`

	var carID string
	err := r.conn.QueryRow(query, stateNumber).Scan(&carID)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return carID, nil
}

func (r *Repository) GetCarsList(userID string, offset int, limit int) ([]models.Car, error) {
	query := `
        SELECT id, id_company, state_number, brand, id_device, id_unicum, count_axis
        FROM cars
        WHERE id_company = $1
        LIMIT $2 OFFSET $3`

	cars := []models.Car{}

	rows, err := r.conn.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var car models.Car
		if err := rows.Scan(&car.ID, &car.IDCompany, &car.StateNumber, &car.Brand, &car.IDDevice, &car.IDUnicum, &car.CountAxis); err != nil {
			return nil, err
		}
		cars = append(cars, car)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cars, nil
}

func (r *Repository) CreateSensor(sensor models.Sensor) (string, error) {
	query := `
        INSERT INTO sensors (car_id, state_number, count_axis, position, pressure, temperature)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id`

	var sensorID string
	err := r.conn.QueryRow(query, sensor.CarID, sensor.StateNumber, sensor.CountAxis, sensor.Position, sensor.Pressure, sensor.Temperature).Scan(&sensorID)
	if err != nil {
		return "", err
	}

	return sensorID, nil
}

func (r *Repository) CreateBreakage(breakage models.Breakage) (string, error) {
	query := `
        INSERT INTO breakages (car_id, state_number, type, description, datetime)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id`

	var breakageID string
	err := r.conn.QueryRow(query,
		breakage.CarID, breakage.StateNumber, breakage.Type, breakage.Description, breakage.Datetime).
		Scan(&breakageID)
	if err != nil {
		return "", err
	}

	return breakageID, nil
}

func (r *Repository) GetSensorsByCarId(carID string) ([]models.Sensor, error) {
	query := `
        SELECT id, car_id, state_number, count_axis, position, pressure, temperature
        FROM sensors
        WHERE car_id = $1`

	var sensors []models.Sensor

	parsedUUID, err := uuid.Parse(carID)
	if err != nil {
		fmt.Println("Error parsing UUID:", err)
		return []models.Sensor{}, nil
	}

	rows, err := r.conn.Query(query, parsedUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sensor models.Sensor
		if err := rows.Scan(&sensor.ID, &sensor.CarID, &sensor.StateNumber, &sensor.CountAxis, &sensor.Position, &sensor.Pressure, &sensor.Temperature); err != nil {
			return nil, err
		}
		sensors = append(sensors, sensor)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sensors, nil
}

func (r *Repository) UpdateSensor(sensor models.Sensor) (models.Sensor, error) {
	query := `
        UPDATE sensors
        SET car_id = $1, state_number = $2, count_axis = $3, position = $4, pressure = $5, temperature = $6
        WHERE state_number = $7 AND position = $8
        RETURNING *
    `

	var updatedSensor models.Sensor
	err := r.conn.QueryRow(query, sensor.CarID, sensor.StateNumber, sensor.CountAxis, sensor.Position, sensor.Pressure, sensor.Temperature, sensor.StateNumber, sensor.Position).
		Scan(&updatedSensor.ID, &updatedSensor.CarID, &updatedSensor.StateNumber, &updatedSensor.CountAxis, &updatedSensor.Position, &updatedSensor.Pressure, &updatedSensor.Temperature)

	if err != nil {
		return models.Sensor{}, fmt.Errorf("error updating sensor: %w", err)
	}
	return updatedSensor, nil
}

func (r *Repository) ChangeWheel(wheelID string, wheel models.Wheel) error {
	query := `
        UPDATE wheels
        SET id_car = $1, count_axis = $2, position = $3, size = $4, cost = $5, brand = $6, model = $7, mileage = $8, min_temperature = $9, min_pressure = $10, max_temperature = $11, max_pressure = $12
        WHERE id = $13`

	r.conn.QueryRow(query, wheel.IDCar, wheel.AxisNumber, wheel.Position, wheel.Size, wheel.Cost, wheel.Brand, wheel.Model, wheel.Mileage, wheel.MinTemperature, wheel.MinPressure, wheel.MaxTemperature, wheel.MaxPressure, wheel.ID)

	return nil
}

func (r *Repository) GetBreakagesByCarId(carID string) ([]models.Breakage, error) {
	query := `
        SELECT id, car_id, state_number, type, description, datetime
        FROM breakages
        WHERE car_id = $1
    `

	var breakages []models.Breakage

	parsedUUID, err := uuid.Parse(carID)
	if err != nil {
		return nil, fmt.Errorf("error parsing carID '%s' into UUID: %w", carID, err)
	}

	rows, err := r.conn.Query(query, parsedUUID)
	if err != nil {
		return nil, fmt.Errorf("error executing query to get breakages: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var breakage models.Breakage
		if err := rows.Scan(&breakage.ID, &breakage.CarID, &breakage.StateNumber, &breakage.Type, &breakage.Description, &breakage.Datetime); err != nil {
			return nil, fmt.Errorf("error scanning row into breakage: %w", err)
		}
		breakages = append(breakages, breakage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}
	return breakages, nil
}

func (r *Repository) SelectAny(table string, key string, val any) (bool, error) {
	query := fmt.Sprintf("SELECT 1 FROM %s WHERE %s = $1 LIMIT 1", table, key)

	var exists int
	err := r.conn.QueryRow(query, val).Scan(&exists)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("error querying database: %v", err)
	}

	return exists == 1, nil
}
