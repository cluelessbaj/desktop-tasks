package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type TaskItem struct {
	ID        int64  `json:"id"`
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
	DueTime   string `json:"due_time"`
	Reminded  bool   `json:"reminded"`
	CreatedAt int64  `json:"created_at"`
}

type AppData struct {
	Tasks []TaskItem `json:"tasks"`
	Notes string     `json:"notes"`
}

type DataStore struct {
	mu       sync.RWMutex
	filePath string
	data     AppData
}

func NewDataStore() (*DataStore, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.Getenv("HOME") + "/.config"
	}
	appDir := filepath.Join(configDir, "desktop-tasks")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, err
	}

	dataFile := filepath.Join(appDir, "data.json")
	ds := &DataStore{
		filePath: dataFile,
		data: AppData{
			Tasks: make([]TaskItem, 0),
			Notes: "",
		},
	}

	ds.Load()
	return ds, nil
}

func (ds *DataStore) Load() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	fileBytes, err := os.ReadFile(ds.filePath)
	if err != nil {
		return
	}

	var d AppData
	if err := json.Unmarshal(fileBytes, &d); err == nil {
		ds.data = d
	}
}

func (ds *DataStore) Save() error {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	bytes, err := json.MarshalIndent(ds.data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ds.filePath, bytes, 0644)
}

func (ds *DataStore) GetJSON() string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	bytes, err := json.Marshal(ds.data)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

func (ds *DataStore) SetFromJSON(rawJSON string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	var d AppData
	if err := json.Unmarshal([]byte(rawJSON), &d); err != nil {
		return err
	}

	ds.data = d

	bytes, err := json.MarshalIndent(ds.data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ds.filePath, bytes, 0644)
}

func (ds *DataStore) AddTask(text, dueTime string) {
	ds.mu.Lock()
	var maxID int64 = 0
	for _, t := range ds.data.Tasks {
		if t.ID > maxID {
			maxID = t.ID
		}
	}

	ds.data.Tasks = append(ds.data.Tasks, TaskItem{
		ID:        maxID + 1,
		Text:      text,
		Completed: false,
		DueTime:   dueTime,
		Reminded:  false,
	})
	ds.mu.Unlock()

	ds.Save()
}
