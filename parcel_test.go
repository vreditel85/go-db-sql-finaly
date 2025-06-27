package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if id == 0 {
		t.Error("Add() returned id = 0, want non-zero")
	}
	parcel.Number = id // сохраняем присвоенный идентификатор

	// get
	got, err := store.Get(id)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// check fields
	if got.Number != parcel.Number {
		t.Errorf("Number = %v, want %v", got.Number, parcel.Number)
	}
	if got.Client != parcel.Client {
		t.Errorf("Client = %v, want %v", got.Client, parcel.Client)
	}
	if got.Status != parcel.Status {
		t.Errorf("Status = %v, want %v", got.Status, parcel.Status)
	}
	if got.Address != parcel.Address {
		t.Errorf("Address = %v, want %v", got.Address, parcel.Address)
	}
	if got.CreatedAt != parcel.CreatedAt {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, parcel.CreatedAt)
	}

	// delete
	err = store.Delete(id)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// verify deletion
	_, err = store.Get(id)
	if err == nil {
		t.Error("Get() after Delete() returned no error, want error")
	}
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	newAddress := "new test address"

	// add parcel
	id, err := store.Add(parcel)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if id == 0 {
		t.Error("Add() returned id = 0, want non-zero")
	}
	parcel.Number = id // сохраняем присвоенный идентификатор

	// set new address
	err = store.SetAddress(id, newAddress)
	if err != nil {
		t.Fatalf("SetAddress() error = %v", err)
	}

	// get and verify
	updatedParcel, err := store.Get(id)
	if err != nil {
		t.Fatalf("Get() after update error = %v", err)
	}

	// check that address was updated
	if updatedParcel.Address != newAddress {
		t.Errorf("Address = %v, want %v", updatedParcel.Address, newAddress)
	}

	// check other fields didn't change
	if updatedParcel.Number != parcel.Number {
		t.Errorf("Number changed from %v to %v", parcel.Number, updatedParcel.Number)
	}
	if updatedParcel.Client != parcel.Client {
		t.Errorf("Client changed from %v to %v", parcel.Client, updatedParcel.Client)
	}
	if updatedParcel.Status != parcel.Status {
		t.Errorf("Status changed from %v to %v", parcel.Status, updatedParcel.Status)
	}
	if updatedParcel.CreatedAt != parcel.CreatedAt {
		t.Errorf("CreatedAt changed from %v to %v", parcel.CreatedAt, updatedParcel.CreatedAt)
	}

	// cleanup
	err = store.Delete(id)
	if err != nil {
		t.Logf("warning: cleanup failed: %v", err)
	}
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	newStatus := "delivered" // новый тестовый статус

	// add parcel
	id, err := store.Add(parcel)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if id == 0 {
		t.Error("Add() returned id = 0, want non-zero")
	}
	parcel.Number = id // сохраняем присвоенный идентификатор

	// set new status
	err = store.SetStatus(id, newStatus)
	if err != nil {
		t.Fatalf("SetStatus() error = %v", err)
	}

	// get and verify
	updatedParcel, err := store.Get(id)
	if err != nil {
		t.Fatalf("Get() after update error = %v", err)
	}

	// check that status was updated
	if updatedParcel.Status != newStatus {
		t.Errorf("Status = %v, want %v", updatedParcel.Status, newStatus)
	}

	// check other fields didn't change
	if updatedParcel.Number != parcel.Number {
		t.Errorf("Number changed from %v to %v", parcel.Number, updatedParcel.Number)
	}
	if updatedParcel.Client != parcel.Client {
		t.Errorf("Client changed from %v to %v", parcel.Client, updatedParcel.Client)
	}
	if updatedParcel.Address != parcel.Address {
		t.Errorf("Address changed from %v to %v", parcel.Address, updatedParcel.Address)
	}
	if updatedParcel.CreatedAt != parcel.CreatedAt {
		t.Errorf("CreatedAt changed from %v to %v", parcel.CreatedAt, updatedParcel.CreatedAt)
	}

	// cleanup
	err = store.Delete(id)
	if err != nil {
		t.Logf("warning: cleanup failed: %v", err)
	}
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		if err != nil {
			t.Fatalf("Add() error = %v", err)
		}
		if id == 0 {
			t.Error("Add() returned id = 0, want non-zero")
		}

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	if err != nil {
		t.Fatalf("GetByClient() error = %v", err)
	}

	// проверяем количество полученных посылок
	if len(storedParcels) != len(parcels) {
		t.Errorf("got %d parcels, want %d", len(storedParcels), len(parcels))
	}

	// check
	for _, parcel := range storedParcels {
		// проверяем что посылка есть в parcelMap
		original, exists := parcelMap[parcel.Number]
		if !exists {
			t.Errorf("parcel with id %d not found in original parcels", parcel.Number)
			continue
		}

		// проверяем все поля
		if parcel.Client != original.Client {
			t.Errorf("Client mismatch for parcel %d: got %d, want %d",
				parcel.Number, parcel.Client, original.Client)
		}
		if parcel.Status != original.Status {
			t.Errorf("Status mismatch for parcel %d: got %s, want %s",
				parcel.Number, parcel.Status, original.Status)
		}
		if parcel.Address != original.Address {
			t.Errorf("Address mismatch for parcel %d: got %s, want %s",
				parcel.Number, parcel.Address, original.Address)
		}
		if parcel.CreatedAt != original.CreatedAt {
			t.Errorf("CreatedAt mismatch for parcel %d: got %v, want %v",
				parcel.Number, parcel.CreatedAt, original.CreatedAt)
		}
	}

	// cleanup
	for id := range parcelMap {
		err := store.Delete(id)
		if err != nil {
			t.Logf("warning: failed to delete parcel %d: %v", id, err)
		}
	}
}
