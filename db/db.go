package db

import (
	"bytes"
	"encoding/json"
	"github.com/dyrkin/zigbee-steward/logger"
	"github.com/dyrkin/zigbee-steward/model"
	"github.com/natefinch/atomic"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

var log = logger.MustGetLogger("db")

type DeviceStore map[string]*model.Device

type dataStore struct {
	Devices *DeviceStore
}

type Db struct {
	ds       *dataStore
	location string
	rw       *sync.RWMutex
}

type Devices struct {
	db *Db
}

func New(filename string) *Db {
	dbPath, err := filepath.Abs(filename)
	if err != nil {
		log.Fatalf("Can't load database: %s. %s", filename, err)
	}
	newDb := &Db{
		ds: &dataStore{
			Devices: &DeviceStore{},
		},
		location: dbPath,
		rw:       &sync.RWMutex{},
	}
	if !newDb.exists() {
		newDb.write()
	}
	newDb.read()

	return newDb
}

func (db *Db) update(updateFn func()) {
	db.rw.Lock()
	defer db.rw.Unlock()
	updateFn()
	db.write()
}

func (db *Db) read() {
	data, err := ioutil.ReadFile(db.location)
	if err != nil {
		log.Fatalf("Can't read database. %s", err)
	}
	if err = json.Unmarshal(data, db.ds); err != nil {
		log.Fatalf("Can't unmarshal database. %s", err)
	}
}

func (db *Db) write() {
	data, err := json.MarshalIndent(db.ds, "", "    ")
	if err != nil {
		log.Fatalf("Can't marshal database. %s", err)
	}
	if err = atomic.WriteFile(db.location, bytes.NewBuffer(data)); err != nil {
		log.Fatalf("Can't write database. %s", err)
	}
	return
}

func (db *Db) exists() bool {
	_, err := os.Stat(db.location)
	return !os.IsNotExist(err)
}

func (db *Db) Devices() Devices {
	return Devices{db: db}
}

func (devices Devices) Add(device *model.Device) {
	devices.db.update(func() {
		(*devices.db.ds.Devices)[device.IEEEAddress] = device
	})
}

func (devices Devices) Get(ieeeAddress string) (*model.Device, bool) {
	devices.db.rw.RLock()
	defer devices.db.rw.RUnlock()
	device, ok := (*devices.db.ds.Devices)[ieeeAddress]
	return device, ok
}

func (devices Devices) GetByNetworkAddress(networkAddress string) (*model.Device, bool) {
	devices.db.rw.RLock()
	defer devices.db.rw.RUnlock()
	for _, d := range *devices.db.ds.Devices {
		if d.NetworkAddress == networkAddress {
			return d, true
		}
	}
	return nil, false
}

func (devices Devices) Remove(ieeeAddress string) {
	devices.db.update(func() {
		delete(*devices.db.ds.Devices, ieeeAddress)
	})
}

func (devices Devices) Exists(ieeeAddress string) bool {
	devices.db.rw.RLock()
	defer devices.db.rw.RUnlock()
	_, ok := (*devices.db.ds.Devices)[ieeeAddress]
	return ok
}
