package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	rt "github.com/networkbootstrap/ztpmanagercode/roottypes"
)

// WebFuncs is the base struct type for the web functions.
// It was either this or wrap it. Urgh. This is easier :)
type WebFuncs struct {
	cachesend  chan rt.Envelope
	configsend chan rt.Envelope
}

func (w WebFuncs) getHost(c echo.Context) error {
	ip := c.Param("ip")
	req := rt.Envelope{}
	req.CRUD = rt.READHOST
	req.Response = make(chan rt.Envelope, 1)
	req.FixedIP = ip
	w.cachesend <- req
	resp := <-req.Response

	if resp.CRUD == rt.OK {
		return c.JSON(http.StatusOK, resp)
	}
	return c.NoContent(http.StatusNoContent)
}

func (w WebFuncs) getHosts(c echo.Context) error {
	req := rt.Envelope{}
	req.CRUD = rt.READHOSTS
	req.Response = make(chan rt.Envelope, 1)
	w.cachesend <- req
	resp := <-req.Response

	return c.JSON(http.StatusOK, resp.HostList)
}

func (w WebFuncs) deleteHost(c echo.Context) error {
	ip := c.Param("ip")

	// Has to be in this order. On the second layer down, the actual cache entry is removed, leaving us just with files.
	req1 := rt.Envelope{}
	req1.CRUD = rt.DELETE
	req1.Response = make(chan rt.Envelope, 1)
	req1.FixedIP = ip
	w.configsend <- req1
	resp1 := <-req1.Response

	if resp1.CRUD == rt.OK {
		req2 := rt.Envelope{}
		req2.CRUD = rt.DELETE
		req2.FixedIP = ip
		req2.Response = make(chan rt.Envelope, 1)
		w.cachesend <- req2
		resp2 := <-req2.Response

		if resp2.CRUD == rt.OK {
			return c.NoContent(http.StatusAccepted)
		}
	}
	return c.NoContent(http.StatusBadRequest)
}

func (w WebFuncs) createHost(c echo.Context) error {
	h := new(rt.Hosts)
	if err := c.Bind(h); err != nil {
		return err
	}

	req := rt.Envelope{}
	req.CRUD = rt.CREATE
	req.FixedIP = h.FixedIP
	req.CfgFile = h.CfgFile
	req.CfgImage = h.CfgImage
	req.Ethernet = h.Ethernet
	req.HostName = h.HostName
	req.Vendor = strings.ToLower(h.Vendor)

	req.Response = make(chan rt.Envelope, 1)
	w.cachesend <- req
	resp := <-req.Response

	if resp.CRUD == rt.OK {
		return c.JSON(http.StatusOK, h)
	}
	return c.NoContent(http.StatusBadRequest)
}

func (w WebFuncs) updateHost(c echo.Context) error {
	h := new(rt.Hosts)
	if err := c.Bind(h); err != nil {
		return err
	}

	ip := c.Param("ip")

	req := rt.Envelope{}
	req.CRUD = rt.UPDATE
	req.UpdateIP = ip
	req.FixedIP = h.FixedIP
	req.CfgFile = h.CfgFile
	req.CfgImage = h.CfgImage
	req.Ethernet = h.Ethernet
	req.HostName = h.HostName
	req.Response = make(chan rt.Envelope, 1)
	w.cachesend <- req
	resp := <-req.Response

	if resp.CRUD == rt.OK {
		return c.JSON(http.StatusOK, h)
	}
	return c.NoContent(http.StatusBadRequest)
}

func (w WebFuncs) save(c echo.Context) error {
	req := rt.Envelope{}
	req.CRUD = rt.SAVECFG
	req.Response = make(chan rt.Envelope, 1)
	w.configsend <- req
	resp := <-req.Response

	if resp.CRUD == rt.OK {
		return c.NoContent(http.StatusAccepted)
	}
	return c.NoContent(http.StatusBadRequest)
}

// StartCfgAPI starts the JSON config API server...
func StartCfgAPI(cachesend chan rt.Envelope, configsend chan rt.Envelope, port int, httpuser string, httppasswd string) (echoSrv *echo.Echo, err error) {
	echoSrv = echo.New()

	echoSrv.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == httpuser && password == httppasswd {
			return true, nil
		}
		return false, nil
	}))

	echoSrv.Use(middleware.Recover())
	echoSrv.HideBanner = true
	web := WebFuncs{}
	web.cachesend = cachesend
	web.configsend = configsend

	// Routes
	echoSrv.POST("/save", web.save)
	echoSrv.POST("/save/", web.save)
	echoSrv.POST("/hosts", web.createHost)
	echoSrv.GET("/hosts", web.getHosts)
	echoSrv.GET("/hosts/", web.getHosts)
	echoSrv.GET("/hosts/:ip", web.getHost)
	echoSrv.DELETE("/hosts/:ip", web.deleteHost)
	echoSrv.PUT("/hosts/:ip", web.updateHost)
	echoport := fmt.Sprintf(":%v", port)

	// Start server
	go func() {
		if err := echoSrv.Start(echoport); err != nil {
			return
		}
	}()

	return echoSrv, nil
}

// StartStaticAPI starts the file server...
func StartStaticAPI(configfiles string, imagesfiles string, configname string, imagesname string) (echoSrv *echo.Echo, err error) {
	echoSrv = echo.New()
	echoSrv.Use(middleware.Recover())

	cfgprefix := fmt.Sprintf("/%s", configname)
	imgprefix := fmt.Sprintf("/%s", imagesname)

	echoSrv.Static(cfgprefix, configfiles)
	echoSrv.Static(imgprefix, imagesfiles)
	echoSrv.HideBanner = true

	// Start server
	go func() {
		if err := echoSrv.Start(":80"); err != nil {
			return
		}
	}()

	return echoSrv, nil
}
