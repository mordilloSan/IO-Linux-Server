package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/godbus/dbus/v5"
)

func registerUpdateRoutes(router *gin.Engine) {
	system := router.Group("/system", authMiddleware())
	{
		system.GET("/updates", getUpdatesHandler)
	}
}

func getUpdatesHandler(c *gin.Context) {
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Println("Failed to connect to system D-Bus:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "dbus connection failed"})
		return
	}

	obj := conn.Object("org.freedesktop.PackageKit", "/org/freedesktop/PackageKit")
	call := obj.CallWithContext(context.Background(),
		"org.freedesktop.PackageKit.Modify.GetUpdates",
		0, "hide-finished")

	if call.Err != nil {
		log.Println("Failed to call GetUpdates:", call.Err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get updates"})
		return
	}

	var transPath dbus.ObjectPath
	if err := call.Store(&transPath); err != nil {
		log.Println("Failed to parse transaction path:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response from PackageKit"})
		return
	}

	// Listen to update signals from the transaction
	signalChan := make(chan *dbus.Signal, 10)
	conn.Signal(signalChan)
	conn.AddMatchSignal(dbus.WithMatchObjectPath(transPath))

	timeout := time.After(10 * time.Second)
	var updates []gin.H

	for {
		select {
		case sig := <-signalChan:
			switch sig.Name {
			case "org.freedesktop.PackageKit.Transaction.Package":
				var infoType uint32
				var pkgID, summary string
				if err := dbus.Store(sig.Body, &infoType, &pkgID, &summary); err == nil {
					updates = append(updates, gin.H{
						"package": pkgID,
						"summary": summary,
					})
				}
			case "org.freedesktop.PackageKit.Transaction.Finished":
				c.JSON(http.StatusOK, gin.H{
					"updates": updates,
				})
				return
			}
		case <-timeout:
			c.JSON(http.StatusGatewayTimeout, gin.H{
				"error":   "timeout waiting for update info",
				"updates": updates,
			})
			return
		}
	}
}
