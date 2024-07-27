package utils

import (
	"encoding/json"
	"fmt"
	"github.com/gnasnik/titan-workerd-api/core/generated/model"
	"io"
	"net"
	"net/http"
)

var privateIPNets = []string{
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"100.64.0.0/10",
	"fd00::/8",
}

func IsPrivateIP(ip net.IP) bool {
	for _, cidr := range privateIPNets {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}

func GetClientIP(r *http.Request) string {
	reqIP := r.Header.Get("X-Real-IP")
	if reqIP == "" {
		h, _, _ := net.SplitHostPort(r.RemoteAddr)
		reqIP = h
	}
	return reqIP
}

func GetLocationByIP(ip string) (*model.Location, error) {
	resp, err := http.Get(fmt.Sprintf("https://api-test1.container1.titannet.io/api/v2/location?ip=%s", ip))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type response struct {
		Code int         `json:"code"`
		Data interface{} `json:"data"`
	}

	var res response
	if err = json.Unmarshal(bytes, &res); err != nil {
		return nil, err
	}

	locationBytes, err := json.Marshal(res.Data)
	if err != nil {
		return nil, err
	}

	var location model.Location
	err = json.Unmarshal(locationBytes, &location)
	if err != nil {
		return nil, err
	}

	return &location, nil
}
