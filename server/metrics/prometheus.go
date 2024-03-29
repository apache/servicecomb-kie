package metrics

import (
	"context"
	"time"

	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/v2/pkg/metrics"
	"github.com/go-chassis/openlog"

	"github.com/apache/servicecomb-kie/server/datasource"
)

const domain = "default"
const project = "default"

func InitMetric() error {
	err := metrics.CreateGauge(metrics.GaugeOpts{
		Key:    "servicecomb_kie_config_count",
		Help:   "use to show the number of config under a specifical domain and project pair",
		Labels: []string{"domain", "project"},
	})
	if err != nil {
		openlog.Error("init servicecomb_kie_config_count Gauge fail:" + err.Error())
		return err
	}
	reportIntervalstr := archaius.GetString("servicecomb.metrics.interval", "5s")
	reportInterval, _ := time.ParseDuration(reportIntervalstr)
	reportTicker := time.NewTicker(reportInterval)
	go func() {
		for {
			_, ok := <-reportTicker.C
			if !ok {
				return
			}
			getTotalConfigCount(project, domain)
		}
	}()
	return nil
}

func getTotalConfigCount(project, domain string) {
	total, err := datasource.GetBroker().GetKVDao().Total(context.TODO(), project, domain)
	if err != nil {
		openlog.Error("set total config number fail: " + err.Error())
		return
	}
	labels := map[string]string{"domain": domain, "project": project}
	err = metrics.GaugeSet("servicecomb_kie_config_count", float64(total), labels)
	if err != nil {
		openlog.Error("set total config number fail:" + err.Error())
		return
	}
}
