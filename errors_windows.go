package main

import (
	"log"
	"os"

	"golang.org/x/sys/windows"
)

func reportError(err error) {
	log.Print(err)
	title, titleErr := windows.UTF16PtrFromString("Notomat")
	message, messageErr := windows.UTF16PtrFromString(err.Error())
	if titleErr == nil && messageErr == nil {
		if _, dialogErr := windows.MessageBox(0, message, title, windows.MB_OK|windows.MB_ICONERROR); dialogErr != nil {
			log.Printf("show error dialog: %v", dialogErr)
		}
	} else {
		log.Printf("prepare error dialog: title=%v, message=%v", titleErr, messageErr)
	}
	os.Exit(1)
}
