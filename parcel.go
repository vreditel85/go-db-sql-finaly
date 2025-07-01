package main

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// Выполняем SQL-запрос на вставку данных
	var result, err = s.db.Exec(
		"INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)",
		p.Client,
		p.Status,
		p.Address,
		p.CreatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert parcel: %w", err)
	}

	// Получаем ID последней вставленной записи
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	p := Parcel{}

	// Выполняем SQL-запрос для выборки строки по number
	row := s.db.QueryRow(
		"SELECT number, client, status, address, created_at FROM parcel WHERE number = ?",
		number,
	)

	// Сканируем данные из строки в структуру Parcel
	err := row.Scan(
		&p.Number,
		&p.Client,
		&p.Status,
		&p.Address,
		&p.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Если запись не найдена, возвращаем пустую структуру и ошибку
			return Parcel{}, fmt.Errorf("parcel with number %d not found", number)
		}
		// В случае других ошибок возвращаем их
		return Parcel{}, fmt.Errorf("failed to get parcel: %w", err)
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	var parcels []Parcel

	// Выполняем SQL-запрос для выборки всех строк по client
	rows, err := s.db.Query(
		"SELECT number, client, status, address, created_at FROM parcel WHERE client = ?",
		client,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query parcels: %w", err)
	}
	defer rows.Close()

	// Итерируем по результатам запроса
	for rows.Next() {
		var p Parcel
		err := rows.Scan(
			&p.Number,
			&p.Client,
			&p.Status,
			&p.Address,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan parcel: %w", err)
		}
		parcels = append(parcels, p)
	}

	// Проверяем ошибки, которые могли возникнуть во время итерации
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return parcels, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// Выполняем SQL-запрос для обновления статуса
	result, err := s.db.Exec(
		"UPDATE parcel SET status = ? WHERE number = ?",
		status,
		number,
	)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Проверяем, была ли обновлена хотя бы одна строка
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("parcel with number %d not found", number)
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// Выполняем обновление с проверкой статуса в одном запросе
	result, err := s.db.Exec(
		"UPDATE parcel SET address = ? WHERE number = ? AND status = 'registered'",
		address,
		number,
	)
	if err != nil {
		return fmt.Errorf("failed to update address: %w", err)
	}

	// Проверяем, была ли обновлена какая-либо строка
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		// Проверяем, существует ли вообще посылка с таким номером
		var exists bool
		err = s.db.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM parcel WHERE number = ?)",
			number,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check parcel existence: %w", err)
		}

		if !exists {
			return fmt.Errorf("parcel with number %d not found", number)
		}
		return fmt.Errorf("address can only be changed for parcels with 'registered' status")
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	result, err := s.db.Exec(
		"DELETE FROM parcel WHERE number = ? AND status = 'registered'",
		number,
	)
	if err != nil {
		return fmt.Errorf("failed to delete parcel: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("parcel not found or status not 'registered'")
	}

	return nil
}
