package metric

import (
	"context"
	"time"

	"github.com/go-chassis/go-chassis/v2/pkg/metrics"
	"github.com/go-chassis/openlog"

	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/datasource"
)

const ReportInterval = 15

func InitMetric(m config.MetricObject) error {
	err := metrics.CreateGauge(metrics.GaugeOpts{
		Key:    "servicecomb_kie_config_count",
		Help:   "use to show the number of config under a specifical domain and project pair",
		Labels: []string{"domain", "project"},
	})
	if err != nil {
		openlog.Error("init servicecomb_kie_config_count Gauge fail:" + err.Error())
		return err
	}
	if m.Domain == "" {
		m.Domain = "default"
	}
	if m.Project == "" {
		m.Project = "default"
	}
	ReportTicker := time.NewTicker(ReportInterval * time.Second)
	go func() {
		for {
			_, ok := <-ReportTicker.C
			if !ok {
				return
			}
			getTotalConfigCount(m.Project, m.Domain)
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
