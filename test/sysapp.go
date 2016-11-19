package main

import (
    "flag"
    "github.com/AnimationMentor/systray"
    "github.com/golang/glog"
    "os"
    "path"
    "path/filepath"
    "runtime"
)

func main() {
    //// Let us tap into the glog flags (like -alsologtostderr=true -log_dir=".")
    //
    
    var iconDirFlag = flag.String("icon_path", ".", "Path to icons files")
    var iconNameFlag = flag.String("icon_name", "sysapp.ico", "Name of icon to show")
    flag.Parse()

    var iconDir string
    if filepath.IsAbs(*iconDirFlag) {
        iconDir = *iconDirFlag
    } else {
        // Default our icon path to be relative to the exe
        thisExe := os.Args[0]
        thisDir := path.Dir(thisExe)
        iconDir = path.Join(thisDir, *iconDirFlag)
    }

    tray := systray.New(iconDir, ".")

    //// Set some test menu items
    items := make([]systray.CallbackInfo, 0, 10)
    items = append(items, systray.CallbackInfo {
        ItemName : "Test Menu 1",
        Callback : func() {
            println("Got menu 1")
        },
    })
    items = append(items, systray.CallbackInfo {
        ItemName : "Test Menu 2",
        Callback : func() {
            println("Got menu 2")
        },
    })
    items = append(items, systray.CallbackInfo {
        ItemName : "Disabled item",
        Callback : func() {
            println("Disabled!!!!")
        },
        Disabled : true,
    })
    items = append(items, systray.CallbackInfo {
        ItemName : "Quit",
        Callback : func() {
            println("Exiting...")
            os.Exit(0)
        },
    })
    tray.AddSystrayMenuItems(items)

    err := tray.Show(*iconNameFlag, "Sysapp Test")
    if err != nil {
        glog.Infoln(err.Error())
    }

    runtime.LockOSThread()
    tray.Run()
    runtime.UnlockOSThread()
}
