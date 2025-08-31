package cache

import (
	"encoding/json"
	"mycelia/errgo"
	"mycelia/str"
	"mycelia/system"
	"os"
	"slices"
	"sync"
)

// -----------------------------------------------------------------------------
// This package is meant for caching highly dynamic runtime data structures,
// such as the broker shape.

// The structures herein are meant to be referenced across packages.
// -----------------------------------------------------------------------------

// -------Broker Routing Structure----------------------------------------------

type channelShape struct {
	Mutex        sync.RWMutex
	Transformers []string
	Subscribers  []string
}

func (cs *channelShape) AddTransformer(address string) {
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()
	if slices.Contains(cs.Transformers, address) {
		return
	}
	cs.Transformers = append(cs.Transformers, address)
}

func (cs *channelShape) RemoveTransformer(address string) {
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()

	for i, transformer := range cs.Transformers {
		if transformer == address {
			cs.Transformers = append(
				cs.Transformers[:i], cs.Transformers[i+1:]...,
			)
			break
		}
	}
}

func (cs *channelShape) AddSubscriber(address string) {
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()
	if slices.Contains(cs.Subscribers, address) {
		return
	}
	cs.Subscribers = append(cs.Subscribers, address)
}

func (cs *channelShape) RemoveSubscriber(address string) {
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()
	for i, subscriber := range cs.Subscribers {
		if subscriber == address {
			cs.Subscribers = append(
				cs.Subscribers[:i], cs.Subscribers[i+1:]...,
			)
			break
		}
	}
}

func newChannelShape() *channelShape {
	return &channelShape{
		Transformers: []string{},
		Subscribers:  []string{},
	}
}

type routeShape struct {
	Mutex    sync.RWMutex
	Channels map[string]*channelShape
}

func (rs *routeShape) Channel(name string) *channelShape {
	rs.Mutex.Lock()
	defer rs.Mutex.Unlock()
	c, exists := rs.Channels[name]
	if !exists {
		c = newChannelShape()
		rs.Channels[name] = c
	}
	return c
}

func newRouteShape() *routeShape {
	return &routeShape{
		Channels: map[string]*channelShape{},
	}
}

// Top node, therefore only one with RWMutex.
type brokerShape struct {
	Routes map[string]*routeShape
	Mutex  sync.RWMutex
}

func NewBrokerShape() *brokerShape {
	return &brokerShape{
		Routes: map[string]*routeShape{},
	}
}

func (bs *brokerShape) Route(name string) *routeShape {
	bs.Mutex.Lock()
	defer bs.Mutex.Unlock()
	r, exists := bs.Routes[name]
	if !exists {
		r = newRouteShape()
		bs.Routes[name] = r
	}
	return r
}

// -----------------------------------------------------------------------------

// The cached broker structure
var BrokerShape *brokerShape = NewBrokerShape()

// -------Broker Routing Structure----------------------------------------------

// Snapshot the shape of the broker + runtime parameters and write it out as a
// shutdown report.
func WriteSnapshot() {
	WriteReport(false)
}

// Writes a report of the current broker statistics, setting the shutdown status
// to fasle.
// gracefulShutdown (bool) is false when snapshotting at runtime and true when
// shutting down expectedly - If broker was shut down unexpectedly it should be
// false becasue it never ran through the shutdown process which flips it to
// true.
func WriteReport(gracefulShutdown bool) {
	reportData := system.NewSystemData()

	reportData.ShutdownReport.GracefulShutdown = &gracefulShutdown

	routeData := errgo.ValueOrPanic(SerializeBrokerShape())
	reportData.Routes = routeData

	jsonData := errgo.ValueOrPanic(json.MarshalIndent(reportData, "", "    "))
	file := errgo.ValueOrPanic(os.Create(system.ShutdownReportFile))
	defer file.Close()

	_, err := file.WriteString(string(jsonData))
	errgo.PanicIfError(err)
	str.ActionPrint("Snapshot ShutdownReport json created in exe directory.")
}

// Boy howdy, is this function ugoly, but necessary.
// The cached data is slightly simpler in shape for simplicity, so there are
// some extra steps taken here to conform it.
func SerializeBrokerShape() (*[]map[string]any, error) {
	bs := BrokerShape
	bs.Mutex.RLock()
	defer bs.Mutex.RUnlock()

	routes := []map[string]any{}
	for rname, rshape := range bs.Routes {
		rshape.Mutex.RLock()
		channels := []map[string]any{}
		for cname, cshape := range rshape.Channels {
			cshape.Mutex.RLock()
			channel := map[string]any{
				"name": cname,
				"transformers": func() []map[string]string {
					out := make([]map[string]string, len(cshape.Transformers))
					for i, addr := range cshape.Transformers {
						out[i] = map[string]string{"address": addr}
					}
					return out
				}(),
				"subscribers": func() []map[string]string {
					out := make([]map[string]string, len(cshape.Subscribers))
					for i, addr := range cshape.Subscribers {
						out[i] = map[string]string{"address": addr}
					}
					return out
				}(),
			}
			cshape.Mutex.RUnlock()
			channels = append(channels, channel)
		}
		rshape.Mutex.RUnlock()
		route := map[string]any{
			"name":     rname,
			"channels": channels,
		}
		routes = append(routes, route)
	}

	return &routes, nil
}
