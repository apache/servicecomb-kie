package notifier

import (
	"sync"
	"time"

	"github.com/apache/servicecomb-kie/server/pubsub"
	"github.com/go-chassis/openlog"
	"github.com/hashicorp/serf/serf"
)

// KVHandler handler serf custom event, it is singleton
type KVHandler struct {
	BatchSize          int
	BatchInterval      time.Duration
	Immediate          bool
	pendingEvents      sync.Map
	pendingEventsCount int
}

func (h *KVHandler) RunFlushTask() {
	for {
		if h.pendingEventsCount >= h.BatchSize {
			h.fireEvents()
		}
		<-time.After(h.BatchInterval)
		h.fireEvents()
	}

}
func (h *KVHandler) HandleEvent(e serf.Event) {
	ue := e.(serf.UserEvent)
	ke, err := pubsub.NewKVChangeEvent(ue.Payload)
	if err != nil {
		openlog.Error("invalid json:" + string(ue.Payload))
	}
	openlog.Debug("kv event:" + ke.Key)
	if h.Immediate { //never retain event, not recommended
		h.FindTopicAndFire(ke)
	} else {
		h.mergeAndSave(ke)
	}

}
func (h *KVHandler) mergeAndSave(ke *pubsub.KVChangeEvent) {
	id := ke.String()
	_, ok := h.pendingEvents.Load(id)
	if ok {
		openlog.Debug("ignore same event: " + id)
		return
	}
	h.pendingEvents.Store(id, ke)
	h.pendingEventsCount++
}
func (h *KVHandler) fireEvents() {
	h.pendingEvents.Range(func(key, value interface{}) bool {
		ke := value.(*pubsub.KVChangeEvent)
		h.FindTopicAndFire(ke)
		h.pendingEvents.Delete(key)
		h.pendingEventsCount--
		return true
	})
}

func (h *KVHandler) FindTopicAndFire(ke *pubsub.KVChangeEvent) {
	topic := pubsub.Topics()
	topic.Range(func(key, value interface{}) bool { //range all topics
		t, err := pubsub.ParseTopic(key.(string))
		if err != nil {
			openlog.Error("can not parse topic " + key.(string) + ": " + err.Error())
			return true
		}
		if t.Match(ke) {
			notifyAndRemoveObservers(value, ke)
		}
		return true
	})
}

func notifyAndRemoveObservers(value interface{}, ke *pubsub.KVChangeEvent) {
	observers := value.(*sync.Map)
	observers.Range(func(id, value interface{}) bool {
		observer := value.(*pubsub.Observer)
		observer.Event <- ke
		observers.Delete(id)
		return true
	})
}
func init() {
	h := &KVHandler{
		BatchInterval: pubsub.DefaultEventBatchInterval,
		BatchSize:     pubsub.DefaultEventBatchSize,
		Immediate:     true,
	}
	pubsub.RegisterHandler(pubsub.EventKVChange, h)
	go h.RunFlushTask()
}
