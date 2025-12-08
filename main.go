package FritzBox_LocalRedirect

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MrEAlderson/FritzBox_LocalRedirect/pkg/avm"
)

type FritzIps struct {
	v4          net.IP
	v6          net.IP
	v6Prefix    *net.IPNet
	refreshTime time.Time
}

func (ips FritzIps) any_match(ip net.IP) bool {
	if ips.v4 != nil && ips.v4.Equal(ip) {
		return true
	}
	if ips.v6 != nil && ips.v6.Equal(ip) {
		return true
	}

	if ips.v6Prefix != nil {
		ipMasked := ip.Mask(ips.v6Prefix.Mask)

		return ips.v6Prefix.IP.Equal(ipMasked)
	}

	return false
}

func (ips FritzIps) all_nil() bool {
	return ips.v4 == nil && ips.v6 == nil && ips.v6Prefix == nil
}

// Config the plugin configuration.
type Config struct {
	FritzURL    string
	RefreshTime string
	TimeoutTime string
	LocalHost   string
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		FritzURL:    "http://192.168.178.1:49000",
		RefreshTime: "30s",
		TimeoutTime: "5s",
		LocalHost:   "my-server:123",
	}
}

// Demo a Demo plugin.
type LRPlugin struct {
	next         http.Handler
	refreshTime  time.Duration
	fritzFetcher avm.FritzBox
	localHost    url.URL
	fritzIps     *FritzIps
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	refreshTime, _ := time.ParseDuration(config.RefreshTime)
	timeoutDuration, _ := time.ParseDuration(config.TimeoutTime)
	rootLogger := slog.Default()

	fritzbox := avm.FritzBox{
		Url:     config.FritzURL,
		Timeout: timeoutDuration,
		Logger:  rootLogger,
	}
	localHost, _ := url.Parse(config.LocalHost)

	return &LRPlugin{
		next:         next,
		refreshTime:  refreshTime,
		fritzFetcher: fritzbox,
		localHost:    *localHost,
		fritzIps:     nil,
	}, nil
}

func FetchIps(a *LRPlugin) {
	v4, _ := a.fritzFetcher.GetWanIpv4()
	v6, _ := a.fritzFetcher.GetwanIpv6()
	v6Prefix, _ := a.fritzFetcher.GetIpv6Prefix()
	fritzIps := &FritzIps{
		v4:          v4,
		v6:          v6,
		v6Prefix:    v6Prefix,
		refreshTime: time.Now(),
	}

	if !fritzIps.all_nil() {
		a.fritzIps = fritzIps
	}
}

func (a *LRPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// force refresh now
	if a.fritzIps == nil {
		FetchIps(a)
		fmt.Println("Fetched fritzIps")

		// Enqueue refresh
	} else {
		now := time.Now()
		last := a.fritzIps.refreshTime

		if now.After(last.Add(a.refreshTime)) {
			go FetchIps(a)
			fmt.Println("Async fritzIps")
		}
	}

	// No fritz info, not worth the effort
	if a.fritzIps == nil {
		a.next.ServeHTTP(rw, req)
		fmt.Println("No fritzIps")
		return
	}

	// Get our ip
	ip := req.Header.Get("X-Forwarded-For")
	var ips []string

	if ip == "" {
		ip = req.Header.Get("X-Real-IP")
	}

	if ip == "" {
		ip = req.Header.Get("CF-Connecting-IP")
	}

	if ip == "" {
		ip = req.RemoteAddr
		// remove port
		pieces := strings.Split(ip, ":")
		ipBlocks := pieces[0 : len(pieces)-1]
		ip = strings.Join(ipBlocks, ":")
		ips = []string{ip}

	} else {
		ips = strings.Split(ip, ",")

		for index, ip := range ips {
			ips[index] = strings.TrimSpace(ip)
		}
	}

	// Redirect?
	for _, rawIp := range ips {
		ip := net.ParseIP(rawIp)
		fmt.Println("Local: ", rawIp)
		fmt.Println("Fritz: ", a.fritzIps)

		if ip != nil && a.fritzIps.any_match(ip) {
			// Redirect!
			url := req.URL
			url.Host = a.localHost.Host

			if a.localHost.Scheme != "" {
				url.Scheme = a.localHost.Scheme
			}

			http.Redirect(rw, req, url.String(), http.StatusTemporaryRedirect)
			fmt.Println("Redirect")
			return
		}
	}
	fmt.Println("No Redirect")
	// Nope, remote access :)
	a.next.ServeHTTP(rw, req)
}
