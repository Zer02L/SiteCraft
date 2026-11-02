package data

import (
	"log"

	"sitecraft/internal/cache"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

// RegisterCacheHooks sets up event listeners to invalidate in-memory cache upon admin updates
func RegisterCacheHooks(app core.App) {
	collections := []string{"site_settings", "site_content", "services", "page_blocks", "articles"}
	app.OnRecordAfterCreateSuccess(collections...).BindFunc(func(e *core.RecordEvent) error {
		invalidateByCollection(e.Record.Collection().Name)
		return e.Next()
	})
	app.OnRecordAfterUpdateSuccess(collections...).BindFunc(func(e *core.RecordEvent) error {
		invalidateByCollection(e.Record.Collection().Name)
		return e.Next()
	})
	app.OnRecordAfterDeleteSuccess(collections...).BindFunc(func(e *core.RecordEvent) error {
		invalidateByCollection(e.Record.Collection().Name)
		return e.Next()
	})
}

func invalidateByCollection(colName string) {
	c := cache.GetStore()
	switch colName {
	case "site_settings":
		c.InvalidateSettings()
		log.Println("[Cache] site_settings invalidated")
	case "site_content":
		c.InvalidateContent()
		log.Println("[Cache] site_content invalidated")
	case "services":
		c.InvalidateServices()
		log.Println("[Cache] services invalidated")
	case "page_blocks":
		c.InvalidateBlocks()
		log.Println("[Cache] page_blocks invalidated")
	case "articles":
		c.InvalidateArticles()
		log.Println("[Cache] articles invalidated")
	}
}
