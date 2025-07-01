package main

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
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
	if assert.NoError(t, err, "Add() should not return error") {
		if assert.NotZero(t, id, "Add() should return non-zero ID") {
			parcel.Number = id // сохраняем присвоенный идентификатор
		}
	}

	// get
	got, err := store.Get(id)
	if assert.NoError(t, err, "Get() should not return error") {
		// check fields
		assert.Equal(t, parcel.Number, got.Number, "Number should match")
		assert.Equal(t, parcel.Client, got.Client, "Client should match")
		assert.Equal(t, parcel.Status, got.Status, "Status should match")
		assert.Equal(t, parcel.Address, got.Address, "Address should match")
		assert.Equal(t, parcel.CreatedAt, got.CreatedAt, "CreatedAt should match")
	}

	// delete
	err = store.Delete(id)
	assert.NoError(t, err, "Delete() should not return error")

	// verify deletion
	_, err = store.Get(id)
	assert.Error(t, err, "Get() after Delete() should return error")
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
