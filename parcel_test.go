package main

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err, "failed to open database")
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err, "Add() should not return error")
	assert.NotZero(t, id, "Add() should return non-zero ID")
	parcel.Number = id // сохраняем присвоенный идентификатор

	// get
	got, err := store.Get(id)
	require.NoError(t, err, "Get() should not return error")
	require.Equal(t, parcel, got, "Get() should return the same parcel")

	// delete
	err = store.Delete(id)
	require.NoError(t, err, "Delete() should not return error")

	// verify deletion
	_, err = store.Get(id)
	assert.Error(t, err, "Get() after Delete() should return error")
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "failed to open database")
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	newAddress := "new test address"

	// add parcel
	id, err := store.Add(parcel)
	require.NoError(t, err, "Add() error")
	assert.NotZero(t, id, "Add() should return non-zero ID")

	parcel.Number = id // сохраняем присвоенный идентификатор

	// set new address
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err, "SetAddress() error")

	// get and verify
	updatedParcel, err := store.Get(id)
	require.NoError(t, err, "Get() after update error")
	require.Equal(t, updatedParcel.Address, newAddress)

	err = store.Delete(id)
	require.NoError(t, err)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "failed to open database")
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()
	newStatus := "delivered" // новый тестовый статус

	// add parcel
	id, err := store.Add(parcel)
	require.NoError(t, err, "Add() error")
	assert.NotZero(t, id, "Add() should return non-zero ID")
	parcel.Number = id // сохраняем присвоенный идентификатор

	// set new status
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err, "SetStatus() error")

	// get and verify
	updatedParcel, err := store.Get(id)
	require.NoError(t, err, "Get() error")

	// check that status was updated
	require.Equal(t, updatedParcel.Status, newStatus)

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
	require.NoError(t, err, "failed to open database")
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
		require.NoError(t, err, "Add() error")
		assert.NotZero(t, id, "Add() should return non-zero ID")

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err, "GetByClient() error")

	// проверяем количество полученных посылок
	require.Len(t, parcels, len(storedParcels))

	// check
	for _, parcel := range storedParcels {
		// проверяем что посылка есть в parcelMap
		original, exists := parcelMap[parcel.Number]
		require.True(t, exists)

		require.Equal(t, original, parcel)
	}

	// cleanup
	for id := range parcelMap {
		err = store.Delete(id)
		assert.NoError(t, err, "warning: failed to delete parcel")
	}
}
