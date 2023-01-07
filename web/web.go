package web

import (
	"fmt"
	"io/fs"
	"log"

	"github.com/MarekWojt/gertdns/auth"
	"github.com/MarekWojt/gertdns/config"
	"github.com/MarekWojt/gertdns/dns"
	"github.com/MarekWojt/gertdns/util"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

var (
	enableDebugMode bool
	r               *router.Router
	ipv4Param       = []byte("ipv4")
	ipv6Param       = []byte("ipv6")
	userParam       = []byte("user")
	passwordParam   = []byte("password")
)

func Init(debugMode bool) {
	enableDebugMode = debugMode
	r = router.New()
	r.GET("/", index)
	r.GET("/update/{domain}/v4", authenticatedRequest(updateV4))
	r.GET("/update/{domain}/v6", authenticatedRequest(updateV6))
	r.GET("/update/{domain}", authenticatedRequest(update))
}

func RunSocket() error {
	httpConfig := config.Config.HTTP
	if httpConfig.Socket != "" {
		log.Printf("Starting HTTP socket in %s with permission %d\n", httpConfig.Socket, httpConfig.SocketFileMode)
		err := fasthttp.ListenAndServeUNIX(httpConfig.Socket, fs.FileMode(httpConfig.SocketFileMode), r.Handler)
		if err != nil {
			return err
		}
	}

	return nil
}

func RunHTTP() error {
	httpConfig := config.Config.HTTP
	if httpConfig.Host != "" {
		log.Printf("Starting HTTP server on %s:%d\n", httpConfig.Host, httpConfig.Port)
		err := fasthttp.ListenAndServe(fmt.Sprintf("%s:%d", httpConfig.Host, httpConfig.Port), r.Handler)
		if err != nil {
			return err
		}
	}

	return nil
}

func index(ctx *fasthttp.RequestCtx) {
	if !enableDebugMode {
		ctx.WriteString("Working")
	} else {
		domains := dns.Get()
		ctx.SetContentType("text/html")
		ctx.WriteString("<!DOCTYPE html><html><head><meta charset='utf-8'><title>gertdns</title></head><body><table><tr><th>Type</th><th>Domain</th><th>IP</th><th>Root domain</th></tr>")
		for _, currentDomain := range domains {
			currentDomain.Mutv4.RLock()
			copiedIpv4s := currentDomain.Ipv4
			currentDomain.Mutv4.RUnlock()

			for name, ip := range copiedIpv4s {
				ctx.WriteString(fmt.Sprintf("<tr><td>A</td><td>%s</td><td>%s</td><td>%s</td><tr>", name, ip, currentDomain.Root))
			}

			currentDomain.Mutv6.RLock()
			copiedIpv6s := currentDomain.Ipv6
			currentDomain.Mutv6.RUnlock()

			for name, ip := range copiedIpv6s {
				ctx.WriteString(fmt.Sprintf("<tr><td>AAAA</td><td>%s</td><td>%s</td><td>%s</td><tr>", name, ip, currentDomain.Root))
			}
		}
		ctx.WriteString("</table></body></html>")
	}
}

func updateV4(ctx *fasthttp.RequestCtx) {
	domain := ctx.UserValue("domain").(string)
	domain = util.ParseDomain(domain)
	ipv4 := string(ctx.QueryArgs().PeekBytes(ipv4Param))
	if ipv4 == "" {
		ctx.WriteString("Missing ipv4 query parameter")
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	err := dns.UpdateIpv4(domain, ipv4)
	if err != nil {
		ctx.WriteString(err.Error())
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}

	ctx.WriteString("OK")
}

func updateV6(ctx *fasthttp.RequestCtx) {
	domain := ctx.UserValue("domain").(string)
	domain = util.ParseDomain(domain)
	ipv6 := string(ctx.QueryArgs().PeekBytes(ipv6Param))
	if ipv6 == "" {
		ctx.WriteString("Missing ipv6 query parameter")
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	err := dns.UpdateIpv6(domain, ipv6)
	if err != nil {
		ctx.WriteString(err.Error())
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}

	ctx.WriteString("OK")
}

func update(ctx *fasthttp.RequestCtx) {
	domain := ctx.UserValue("domain").(string)
	domain = util.ParseDomain(domain)

	ipv4 := string(ctx.QueryArgs().PeekBytes(ipv4Param))
	ipv6 := string(ctx.QueryArgs().PeekBytes(ipv6Param))

	if ipv4 == "" && ipv6 == "" {
		ctx.WriteString("Provide at least one these query parameters: ipv4, ipv6")
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	if ipv4 != "" {
		err := dns.UpdateIpv4(domain, ipv4)
		if err != nil {
			ctx.WriteString(err.Error())
			ctx.SetStatusCode(fasthttp.StatusNotFound)
			return
		}
	}

	if ipv6 != "" {
		err := dns.UpdateIpv6(domain, ipv6)
		if err != nil {
			ctx.WriteString(err.Error())
			ctx.SetStatusCode(fasthttp.StatusNotFound)
			return
		}
	}

	ctx.WriteString("OK")
}

func authenticatedRequest(request func(ctx *fasthttp.RequestCtx)) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		domain, ok := ctx.UserValue("domain").(string)
		if !ok {
			ctx.WriteString("Missing domain")
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			return
		}
		domain = util.ParseDomain(domain)

		user := string(ctx.QueryArgs().PeekBytes(userParam))
		if user == "" {
			ctx.WriteString("Missing user query parameter")
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			return
		}
		password := string(ctx.QueryArgs().PeekBytes(passwordParam))
		if user == "" {
			ctx.WriteString("Missing password query parameter")
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			return
		}

		authRequest := auth.PasswordAuthenticationRequest{
			Domain:   domain,
			User:     user,
			Password: password,
		}

		authenticated, err := auth.IsPasswordAuthenticated(authRequest)
		if err != nil && !authenticated {
			ctx.WriteString("Authentication failed: " + err.Error())
			ctx.SetStatusCode(fasthttp.StatusForbidden)
			return
		}

		if err != nil {
			ctx.WriteString("Internal server error")
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}

		if !authenticated {
			ctx.WriteString("Authentication failed")
			ctx.SetStatusCode(fasthttp.StatusForbidden)
			return
		}

		request(ctx)
	}
}
