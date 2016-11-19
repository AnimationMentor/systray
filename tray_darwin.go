package systray

import (
    "unsafe"
    "path/filepath"
    "runtime"
)

/*
#cgo linux pkg-config: gtk+-3.0
#cgo linux CFLAGS: -DLINUX
#cgo windows CFLAGS: -DWIN32
#cgo darwin CFLAGS: -DDARWIN -x objective-c
#cgo darwin LDFLAGS: -framework Cocoa
#include <stdlib.h>
void runApplication(const char *title, const char *initialIcon, const char *initialHint, void *manager);
void stopApplication(void);
void addSystrayMenuItem(const char *item, void *, unsigned int index, unsigned char enabled, unsigned char checked);
void clearSystrayMenuItems(void);
void setIcon(const char *path);
void setHint(const char *hint);
*/
import "C"

func (p *_Systray) Stop() error {
    C.stopApplication()
    return nil
}

func (p *_Systray) Show(file string, hint string) error {
    err := p.SetIcon(file)
    if err != nil {
        return err
    }
    err = p.SetTooltip(hint)
    if err != nil {
        return err
    }
    return p.SetVisible(true)
}

func (p *_Systray) OnClick(fun func()) {
    p.lclick = fun
    p.rclick = fun
    p.dclick = fun
}

func (p *_Systray) SetTooltip(tooltip string) error {
    p.currentHint = tooltip
    if p.isCreated {
        cTooltip := C.CString(tooltip)
        defer C.free(unsafe.Pointer(cTooltip))
        C.setHint(cTooltip)
    }
    return nil
}

func (p *_Systray) SetIcon(file string) error {
    p.currentIcon = file
    if p.isCreated {
        cFile := C.CString(file)
        defer C.free(unsafe.Pointer(cFile))
        C.setIcon(cFile)
    }
    return nil
}

func (p *_Systray) SetVisible(visible bool) error {
    // Does this have any meaning for darwin?
    return nil
}

func (p *_Systray) Run() error {
    cTitle := C.CString(p.currentHint)
    defer C.free(unsafe.Pointer(cTitle))
    
    cIconPath := C.CString(filepath.Join(p.iconPath, p.currentIcon))
    defer C.free(unsafe.Pointer(cIconPath))

    println("Running main loop on systray", p)

    // Enter the main loop - this calls [NSApplication run] internally, which *must*
    // execute on the main thread.
    // We call LockOSThread() here just in case, but, really, call it earlier!
    runtime.LockOSThread()
    C.runApplication(cTitle, cIconPath, cTitle, unsafe.Pointer(p))
    runtime.UnlockOSThread()
    
    // If reached, user clicked Exit
    p.isExiting = true

    return nil
}

func _NewSystray(iconPath string, clientPath string) *_Systray {
    tray, err := _NewSystrayEx(iconPath)
    if err != nil {
        panic(err)
    }
    return tray
}

func _NewSystrayEx(iconPath string) (*_Systray, error) {
    ni := &_Systray{iconPath, "", "", false, false, make([]CallbackInfo, 0, 10), func() {}, func() {}, func() {}}
    return ni, nil
}

type _Systray struct {
    iconPath          string
    currentIcon       string
    currentHint       string
    isExiting         bool
    isCreated         bool
    menuItemCallbacks []CallbackInfo
    lclick            func()
    rclick            func()
    dclick            func()
}


func (p *_Systray) insertMenuItem(info CallbackInfo, index int) {
    // TODO - insert item into array at desired index
    p.menuItemCallbacks = append(p.menuItemCallbacks, info)
}

func (p *_Systray) appendMenuItem(info CallbackInfo) {
    index := len(p.menuItemCallbacks)
    p.menuItemCallbacks = append(p.menuItemCallbacks, info)
    if p.isCreated {
        p.addItemToNativeMenu(info, index)
    }
}

func (p *_Systray) addItemToNativeMenu(info CallbackInfo, index int) {
    cItemName := C.CString(info.ItemName)
    defer C.free(unsafe.Pointer(cItemName))
    cIndex := C.uint(index)
    var cEnabled C.uchar
    if info.Disabled || info.Callback == nil {
        cEnabled = C.uchar(0)
    } else {
        cEnabled = C.uchar(1)
    }
    var cChecked C.uchar
    if info.Checked {
        cChecked = C.uchar(1)
    } else {
        cChecked = C.uchar(0)
    }
    C.addSystrayMenuItem(cItemName, unsafe.Pointer(p), cIndex, cEnabled, cChecked)
}

func (p *_Systray) AddSystrayMenuItems(items []CallbackInfo) {
    for _, info := range items {
        p.appendMenuItem(info)
    }
}

func (p *_Systray) ClearSystrayMenuItems() {
    p.menuItemCallbacks = make([]CallbackInfo, 0, 10)
    if p.isCreated {
        C.clearSystrayMenuItems()
    } 
}

func (p *_Systray) handleMenuClick(index int) {
    println("Want to handle menu click for index", index)
    if index >= 0 && index < len(p.menuItemCallbacks) {
        p.menuItemCallbacks[index].Callback()
    }
}

/*
 * C API to provide hooks back into Go. Without the ability to pass Go function
 * pointers into C functions, the C code needs to know a priori about these
 * hooks.
 */
//export menuClickCallback
func menuClickCallback(manager unsafe.Pointer, index int) {
    if manager != nil {
        p := (*_Systray)(manager)
        p.handleMenuClick(index)
    }
}

//export menuCreatedCallback
func menuCreatedCallback(manager unsafe.Pointer) {
    if manager != nil {
        p := (*_Systray)(manager)
        p.isCreated = true
        // Add all previously registered callbacks to the menu
        for idx, info := range p.menuItemCallbacks {
            println("Adding callback for", info.ItemName)
            p.addItemToNativeMenu(info, idx)
        }
        p.SetTooltip(p.currentHint)
    }
}

