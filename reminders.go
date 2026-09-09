package main

import (
	"fmt"
	"os/exec"
	"time"
)

func startReminderTicker(ds *DataStore) {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			checkReminders(ds)
		}
	}()
}

func checkReminders(ds *DataStore) {
	now := time.Now()
	currentHM := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())

	ds.mu.Lock()
	changed := false
	for i := range ds.data.Tasks {
		task := &ds.data.Tasks[i]
		if !task.Completed && !task.Reminded && task.DueTime != "" {
			if currentHM >= task.DueTime {
				sendDesktopNotification("Desktop Task Due", task.Text)
				task.Reminded = true
				changed = true
			}
		}
	}
	ds.mu.Unlock()

	if changed {
		ds.Save()
	}
}

func sendDesktopNotification(title, message string) {
	cmd := exec.Command("notify-send", title, message, "-i", "appointment-soon", "-u", "normal")
	_ = cmd.Run()
}
