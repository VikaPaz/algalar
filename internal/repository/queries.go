package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/VikaPaz/algalar/internal/models"
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

func (r *Repository) GetIDByEmailAndPassword(email, password string) (string, error) {
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

// func (r *Repository) CreateCar(car models.Car) (string, error) {
// 	query := `
//         INSERT INTO cars (id_company, state_namber, brand, id_device, id_unicum, count_axis)
//         VALUES ($1, $2, $3, $4, $5, $6)
//         RETURNING id`

// 	var carID string
// 	err := r.conn.QueryRow(query, car.IDCompany, car.StateNamber, car.Brand, car.IDDevice, car.IDUnicum, car.CountAxis).Scan(&carID)
// 	if err != nil {
// 		return "", err
// 	}

// 	return carID, nil
// }

func (r *Repository) CreateCar(car models.Car) (string, error) {
	query := `
        INSERT INTO cars (state_namber, brand, id_device, id_unicum, count_axis)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id`

	var carID string
	err := r.conn.QueryRow(query, car.StateNumber, car.Brand, car.IDDevice, car.IDUnicum, car.CountAxis).Scan(&carID)
	if err != nil {
		return "", err
	}

	return carID, nil
}

func (r *Repository) CreateWheel(wheel models.Wheel) (string, error) {
	fmt.Println(wheel)
	query := `
        INSERT INTO wheels (id_car, axis_number, position, size, cost, brand, model, mileage, min_temperature, min_pressure, max_temperature, max_pressure)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        RETURNING id`

	var wheelID string
	err := r.conn.QueryRow(query, wheel.IDCar, wheel.AxisNumber, wheel.Position, wheel.Size, wheel.Cost, wheel.Brand, wheel.Model, wheel.Mileage, wheel.MinTemperature, wheel.MinPressure, wheel.MaxTemperature, wheel.MaxPressure).Scan(&wheelID)
	if err != nil {
		return "", err
	}

	return wheelID, nil
}

func (r *Repository) GetWheelById(wheelID string) (models.Wheel, error) {
	query := `
        SELECT id_car, axis_number, position, size, cost, brand, model, mileage, min_temperature, min_pressure, max_temperature, max_pressure
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

func (r *Repository) ChangeWheel(wheelID string, wheel models.Wheel) error {
	query := `
        UPDATE wheels
        SET id_car = $1, axis_number = $2, position = $3, size = $4, cost = $5, brand = $6, model = $7, mileage = $8, min_temperature = $9, min_pressure = $10, max_temperature = $11, max_pressure = $12
        WHERE id = $13`

	_, err := r.conn.Exec(query, wheel.IDCar, wheel.AxisNumber, wheel.Position, wheel.Size, wheel.Cost, wheel.Brand, wheel.Model, wheel.Mileage, wheel.MinTemperature, wheel.MinPressure, wheel.MaxTemperature, wheel.MaxPressure, wheelID)
	if err != nil {
		return err
	}

	return nil
}
