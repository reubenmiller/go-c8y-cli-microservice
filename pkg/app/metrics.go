package app

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	//
	// Application Information for prometheus
	//
	applicationInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "app_info",
			Help: "Application information like version and other static information",
		},
		[]string{"version", "branch", "commit", "buildTime"},
	)

	// Version application version number
	Version string

	// Branch is the repository branch which was used to build the application
	Branch string

	// Commit number
	Commit string

	// BuildTime is the time which the application was built
	BuildTime string
)

func init() {
	applicationInfo.WithLabelValues(Version, Branch, Commit, BuildTime).Set(1)
}
