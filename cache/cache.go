// Package cache provides a cache for CRUD actions on hosts
package cache

import (
	"bytes"
	"sync"

	"github.com/BurntSushi/toml"
	rt "github.com/networkbootstrap/ztpmanagercode/roottypes"
)

// Create creates and runs a CRUD based cache
//func Create(cache map[string]rt.Hosts, wg *sync.WaitGroup, recvch chan rt.Envelope, finish chan struct{}) {
func Create(cache map[string]rt.Hosts, wg *sync.WaitGroup) (recvch chan rt.Envelope, finish chan struct{}) {
	recvch = make(chan rt.Envelope, 1)
	finish = make(chan struct{})

	go func() {
		locked := false
		for {
			select {
			case recv := <-recvch:

				switch recv.CRUD {
				case rt.CREATE:
					if !locked {
						insert := rt.Hosts{}
						insert.CfgFile = recv.CfgFile
						insert.CfgImage = recv.CfgImage
						insert.Ethernet = recv.Ethernet
						insert.FixedIP = recv.FixedIP
						insert.HostName = recv.HostName
						insert.Vendor = recv.Vendor
						// Insert
						cache[insert.FixedIP] = insert
						// Now check
						if _, ok := cache[insert.FixedIP]; ok {
							resp := rt.Envelope{}
							resp.Response = recv.Response
							resp.CRUD = rt.OK
							resp.Hosts = insert
							resp.Response <- resp
						} else {
							resp := rt.Envelope{}
							resp.Response = recv.Response
							resp.CRUD = rt.ERROR
							resp.Hosts = insert
							resp.Response <- resp
						}
					}

				case rt.READHOST:
					if !locked {
						read := rt.Hosts{}
						read.CfgFile = cache[recv.FixedIP].CfgFile
						read.CfgImage = cache[recv.FixedIP].CfgImage
						read.Ethernet = cache[recv.FixedIP].Ethernet
						read.FixedIP = cache[recv.FixedIP].FixedIP
						read.HostName = cache[recv.FixedIP].HostName
						read.Vendor = cache[recv.FixedIP].Vendor
						if _, ok := cache[read.FixedIP]; ok {
							resp := rt.Envelope{}
							resp.Response = recv.Response
							resp.CRUD = rt.OK
							resp.Hosts = read
							resp.Response <- resp
						} else {
							resp := rt.Envelope{}
							resp.Response = recv.Response
							resp.CRUD = rt.ERROR
							resp.Hosts = read
							resp.Response <- resp
						}
					}
				case rt.READHOSTS:
					if !locked {
						hosts := rt.Envelope{}
						for k := range cache {
							hosts.HostList = append(hosts.HostList, k)
						}
						hosts.Response = recv.Response
						hosts.CRUD = rt.OK
						hosts.Response <- hosts
					}

				case rt.UPDATE:
					if !locked {
						update := rt.Hosts{}
						update.CfgFile = recv.CfgFile
						update.CfgImage = recv.CfgImage
						update.Ethernet = recv.Ethernet
						update.FixedIP = recv.FixedIP
						update.HostName = recv.HostName
						delete(cache, recv.UpdateIP)
						cache[recv.FixedIP] = update

						// Now check
						if _, ok := cache[recv.FixedIP]; ok {
							resp := rt.Envelope{}
							resp.Response = recv.Response
							resp.CRUD = rt.OK
							resp.Response <- resp
						}
					}
				case rt.DELETE:
					if !locked {
						deleted := false
						if _, ok := cache[recv.FixedIP]; ok {
							deleted = true
						}
						delete(cache, recv.FixedIP)
						if _, ok := cache[recv.FixedIP]; !ok && deleted {
							resp := rt.Envelope{}
							resp.Response = recv.Response
							resp.CRUD = rt.OK
							resp.Response <- resp
						} else {
							resp := rt.Envelope{}
							resp.Response = recv.Response
							resp.CRUD = rt.ERROR
							resp.Response <- resp
						}
					}
				case rt.STRING:
					if !locked {
						var b []byte
						buf := bytes.NewBuffer(b)
						tomlEncoder := toml.NewEncoder(buf)
						err := tomlEncoder.Encode(cache)
						if err != nil {
							resp := rt.Envelope{}
							resp.Response = recv.Response
							resp.CRUD = rt.ERROR
							resp.Response <- resp
						}
						resp := rt.Envelope{}
						resp.Response = recv.Response
						resp.CRUD = rt.OK
						resp.String = buf.String()
						resp.Response <- resp
					}
				case rt.LOCK:
					locked = true
					resp := rt.Envelope{}
					resp.Response = recv.Response
					resp.CRUD = rt.OK
					resp.Response <- resp
				case rt.UNLOCK:
					locked = false
					resp := rt.Envelope{}
					resp.Response = recv.Response
					resp.CRUD = rt.OK
					resp.Response <- resp
				}
			case <-finish:
				// We're finished *sob*
				wg.Done()
				return
			}
		}
	}()
	return
}
