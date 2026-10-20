package models

// FleetSite represents a single isolated SiteCraft site instance.
type FleetSite struct {
	SiteID     string `json:"site_id"`
	Preset     string `json:"preset"`
	Theme      string `json:"theme"`
	Port       string `json:"port"`
	AdminEmail string `json:"admin_email"`
	CreatedAt  string `json:"created_at"`
	Status     string `json:"status"` // "running" or "stopped"
	PID        string `json:"pid"`
	URL        string `json:"url"`
	AdminURL   string `json:"admin_url"`
}

// FleetStats represents summary statistics for all fleet instances.
type FleetStats struct {
	TotalSites   int `json:"total_sites"`
	RunningSites int `json:"running_sites"`
	StoppedSites int `json:"stopped_sites"`
}
