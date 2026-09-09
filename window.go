package main

/*
#cgo pkg-config: gtk+-3.0 webkit2gtk-4.1
#include <gtk/gtk.h>
#include <webkit2/webkit2.h>

static gboolean on_key_press_cb(GtkWidget *widget, GdkEventKey *event, gpointer user_data) {
    (void)user_data;
    if (event->keyval == GDK_KEY_Escape) {
        gtk_window_set_keep_above(GTK_WINDOW(widget), FALSE);
        gtk_window_set_keep_below(GTK_WINDOW(widget), TRUE);
        return TRUE;
    }
    return FALSE;
}

static gboolean on_window_state_cb(GtkWidget *widget, GdkEventWindowState *event, gpointer user_data) {
    (void)event; (void)user_data;
    return FALSE;
}

static void configure_desktop_window(void *win_ptr, void *webkit_ptr, int width, int height, int margin_right, int margin_top) {
    GtkWidget *win = GTK_WIDGET(win_ptr);

    gtk_window_set_decorated(GTK_WINDOW(win), FALSE);
    gtk_window_set_skip_taskbar_hint(GTK_WINDOW(win), TRUE);
    gtk_window_set_skip_pager_hint(GTK_WINDOW(win), TRUE);
    gtk_window_stick(GTK_WINDOW(win));
    gtk_window_set_keep_below(GTK_WINDOW(win), TRUE);

    // Transparency
    GdkScreen *screen = gtk_widget_get_screen(win);
    GdkVisual *visual = gdk_screen_get_rgba_visual(screen);
    if (visual && gdk_screen_is_composited(screen)) {
        gtk_widget_set_visual(win, visual);
    }
    gtk_widget_set_app_paintable(win, TRUE);

    if (webkit_ptr) {
        GdkRGBA transparent = {0.0, 0.0, 0.0, 0.0};
        webkit_web_view_set_background_color(WEBKIT_WEB_VIEW(webkit_ptr), &transparent);
    }

    g_signal_connect(win, "key-press-event", G_CALLBACK(on_key_press_cb), NULL);
    g_signal_connect(win, "window-state-event", G_CALLBACK(on_window_state_cb), NULL);

    // Position window
    GdkDisplay *display = gtk_widget_get_display(win);
    GdkMonitor *monitor = gdk_display_get_primary_monitor(display);
    if (!monitor) {
        monitor = gdk_display_get_monitor_at_point(display, 0, 0);
    }

    GdkRectangle workarea = {0, 26, 1920, 1054};
    if (monitor) {
        gdk_monitor_get_workarea(monitor, &workarea);
    }

    int win_h = height;
    if (workarea.height < win_h + 40) {
        win_h = workarea.height - 40;
    }

    int target_x = workarea.x + workarea.width - width - margin_right;
    int target_y = workarea.y + margin_top;

    gtk_widget_set_size_request(win, width, win_h);
    gtk_window_move(GTK_WINDOW(win), target_x, target_y);
}

static void toggle_window_foreground(void *win_ptr, int to_foreground) {
    GtkWindow *win = GTK_WINDOW(win_ptr);
    if (to_foreground) {
        gtk_window_set_keep_below(win, FALSE);
        gtk_window_set_keep_above(win, TRUE);
        gtk_window_present(win);
    } else {
        gtk_window_set_keep_above(win, FALSE);
        gtk_window_set_keep_below(win, TRUE);
    }
}

static void start_window_drag(void *win_ptr) {
    GtkWidget *win = GTK_WIDGET(win_ptr);
    GdkDisplay *display = gtk_widget_get_display(win);
    GdkSeat *seat = gdk_display_get_default_seat(display);
    GdkDevice *device = gdk_seat_get_pointer(seat);

    gint x_root, y_root;
    gdk_device_get_position(device, NULL, &x_root, &y_root);

    gtk_window_begin_move_drag(GTK_WINDOW(win), 1, x_root, y_root, gtk_get_current_event_time());
}
*/
import "C"
import (
	"unsafe"
)

type WindowController struct {
	winPtr    unsafe.Pointer
	webkitPtr unsafe.Pointer
	isFore    bool
}

func NewWindowController(winPtr, webkitPtr unsafe.Pointer) *WindowController {
	wc := &WindowController{
		winPtr:    winPtr,
		webkitPtr: webkitPtr,
		isFore:    false,
	}
	C.configure_desktop_window(winPtr, webkitPtr, C.int(340), C.int(660), C.int(20), C.int(38))
	return wc
}

func (wc *WindowController) ToggleForeground() bool {
	wc.isFore = !wc.isFore
	flag := 0
	if wc.isFore {
		flag = 1
	}
	C.toggle_window_foreground(wc.winPtr, C.int(flag))
	return wc.isFore
}

func (wc *WindowController) SetForeground(fore bool) {
	wc.isFore = fore
	flag := 0
	if wc.isFore {
		flag = 1
	}
	C.toggle_window_foreground(wc.winPtr, C.int(flag))
}

func (wc *WindowController) StartDrag() {
	C.start_window_drag(wc.winPtr)
}
