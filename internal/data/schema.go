package data

import (
	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

// InitSchema ensures site_settings, site_content, page_blocks, services, and leads collections exist.
func InitSchema(app core.App) error {
	if err := initSiteSettingsCollection(app); err != nil {
		return err
	}
	if err := initSiteContentCollection(app); err != nil {
		return err
	}
	if err := initPageBlocksCollection(app); err != nil {
		return err
	}
	if err := initServicesCollection(app); err != nil {
		return err
	}
	if err := initLeadsCollection(app); err != nil {
		return err
	}
	if err := initArticlesCollection(app); err != nil {
		return err
	}
	if err := initServiceOrdersCollection(app); err != nil {
		return err
	}
	if err := initChatMessagesCollection(app); err != nil {
		return err
	}
	if err := initCouponsCollection(app); err != nil {
		return err
	}
	if err := initResourcesCollection(app); err != nil {
		return err
	}
	if err := initReviewsCollection(app); err != nil {
		return err
	}
	if err := initBranchesCollection(app); err != nil {
		return err
	}
	if err := initVacanciesCollection(app); err != nil {
		return err
	}
	RegisterCacheHooks(app)
	return nil
}
