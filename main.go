package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"strings"

	webview "localwebview"
)

//go:embed assets/*
var assetsFS embed.FS

const AppVersion = "2.0.0 (Go + WebKitGTK Synthwave Edition)"

func main() {
	args := os.Args[1:]

	if len(args) > 0 {
		switch args[0] {
		case "--version":
			fmt.Printf("desktop-tasks version %s\n", AppVersion)
			return
		case "--test-notification":
			sendDesktopNotification("Desktop Tasks Test", "Desktop notifications are working perfectly with Go!")
			fmt.Println("Test notification dispatched.")
			return
		case "--toggle", "-t":
			if SendIPCCommand("toggle") {
				return
			}
		case "--notes":
			if SendIPCCommand("notes") {
				return
			}
		case "--tasks":
			if SendIPCCommand("tasks") {
				return
			}
		case "--add":
			if len(args) > 1 {
				cmd := "add " + strings.Join(args[1:], " ")
				if SendIPCCommand(cmd) {
					return
				}
			}
		}
	}

	// Single instance check: if daemon is already running, toggle it
	if SendIPCCommand("toggle") {
		fmt.Println("desktop-tasks daemon is already running. Toggled focus.")
		return
	}

	// Initialize DataStore
	ds, err := NewDataStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing data store: %v\n", err)
		return
	}

	// Start reminder background ticker
	startReminderTicker(ds)

	// Serve assets locally on ephemeral loopback port
	subFS, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load assets: %v\n", err)
		return
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start local server: %v\n", err)
		return
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	go func() {
		_ = http.Serve(listener, http.FileServer(http.FS(subFS)))
	}()

	// Initialize Webview
	w := webview.New(false)
	defer w.Destroy()

	w.SetTitle("Desktop Tasks & Notes")
	w.SetSize(340, 660, webview.HintNone)

	// Window controller for GTK desktop hints, transparency & positioning
	wc := NewWindowController(w.Window(), w.BrowserController())

	// Bind Go callbacks to JavaScript
	_ = w.Bind("backend_loadData", func() string {
		return ds.GetJSON()
	})

	_ = w.Bind("backend_saveData", func(rawJSON string) {
		_ = ds.SetFromJSON(rawJSON)
	})

	_ = w.Bind("backend_startDrag", func() {
		wc.StartDrag()
	})

	_ = w.Bind("backend_toggleForeground", func() {
		wc.ToggleForeground()
	})

	_ = w.Bind("backend_sinkDesktop", func() {
		wc.SetForeground(false)
	})

	// Start IPC Server
	sockPath := getSocketPath()
	err = startIPCServer(sockPath, func(cmd string) {
		w.Dispatch(func() {
			if cmd == "toggle" {
				isFore := wc.ToggleForeground()
				if isFore {
					w.Eval("if (window.onExternalToggle) window.onExternalToggle();")
				}
			} else if cmd == "notes" {
				wc.SetForeground(true)
				w.Eval("if (window.onExternalSwitchTab) window.onExternalSwitchTab('notes');")
			} else if cmd == "tasks" {
				wc.SetForeground(true)
				w.Eval("if (window.onExternalSwitchTab) window.onExternalSwitchTab('tasks');")
			} else if strings.HasPrefix(cmd, "add ") {
				taskText := strings.TrimPrefix(cmd, "add ")
				escaped := strings.ReplaceAll(taskText, "'", "\\'")
				w.Eval(fmt.Sprintf("if (window.onExternalAddTask) window.onExternalAddTask('%s');", escaped))
			}
		})
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to start IPC server: %v\n", err)
	}

	// Navigate to local app
	w.Navigate(fmt.Sprintf("http://127.0.0.1:%d/index.html", port))

	w.Run()
}
