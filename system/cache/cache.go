package cache

import (
	"fmt"
	"mycelia/globals"
	"mycelia/str"
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

// -------Serializers----------------------------------------------

// Boy howdy, is this function ugoly, but necessary.
// The cached data is slightly simpler in shape for simplicity, so there are
// some extra steps taken here to conform it to expected values for config files
// or shutdown reports.
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

func PrintBrokerStructure() {
	if !globals.PrintTree {
		return
	}

	routeExpr := "  | - [route] %s\n"
	channelExpr := "        | - [channel] %s\n"
	transformerExpr := "              | - [transformer] %s\n"
	subscriberExpr := "              | - [subscriber] %s\n"

	str.PrintCenteredHeader("Broker Shape")
	fmt.Println("\n[broker]")

	bs := BrokerShape
	bs.Mutex.RLock()
	for routeName, rshape := range bs.Routes {
		fmt.Printf(routeExpr, routeName)

		rshape.Mutex.RLock()
		for channelName, cshape := range rshape.Channels {
			fmt.Printf(channelExpr, channelName)

			cshape.Mutex.RLock()
			// Print transformers first (to match current format)
			for _, addr := range cshape.Transformers {
				fmt.Printf(transformerExpr, addr)
			}
			// Then print subscribers
			for _, addr := range cshape.Subscribers {
				fmt.Printf(subscriberExpr, addr)
			}
			cshape.Mutex.RUnlock()
		}
		rshape.Mutex.RUnlock()
	}
	bs.Mutex.RUnlock()

	str.PrintAsciiLine()
	fmt.Println()
}
