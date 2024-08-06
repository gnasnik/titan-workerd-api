package api

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-workerd-api/core/generated/model"
	"github.com/gnasnik/titan-workerd-api/core/geo"
	"math"
	"strconv"

	"github.com/golang/geo/s2"
)

type IPCoordinate interface {
	GetLatLng(ctx context.Context, ip string) (float64, float64, error)
}

type ipCoordinate struct {
	// *geoip2.Reader
}

func NewIPCoordinate() IPCoordinate {
	return &ipCoordinate{}
}

func (coordinate *ipCoordinate) GetLatLng(ctx context.Context, ip string) (float64, float64, error) {
	loc, err := geo.GetIpLocation(ctx, ip, model.LanguageEN)
	if err != nil {
		return 0, 0, err
	}

	longitude, err := strconv.ParseFloat(loc.Longitude, 64)
	if err != nil {
		return 0, 0, err
	}

	latitude, err := strconv.ParseFloat(loc.Latitude, 64)
	if err != nil {
		return 0, 0, err
	}

	return latitude, longitude, nil
}

func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	p1 := s2.PointFromLatLng(s2.LatLngFromDegrees(lat1, lon1))
	p2 := s2.PointFromLatLng(s2.LatLngFromDegrees(lat2, lon2))

	distance := s2.ChordAngleBetweenPoints(p1, p2).Angle().Radians()

	distanceKm := distance * 6371.0

	return distanceKm
}

func GetUserNearestIP(ctx context.Context, userIP string, ipList []string, coordinate IPCoordinate) (string, error) {
	lat1, lon1, err := coordinate.GetLatLng(ctx, userIP)
	if err != nil {
		return "", err
	}

	ipDistanceMap := make(map[string]float64)
	for _, ip := range ipList {
		lat2, lon2, err := coordinate.GetLatLng(ctx, ip)
		if err != nil {
			log.Errorf("get %s latLng error %s", ip, err.Error())
			continue
		}

		distance := calculateDistance(lat1, lon1, lat2, lon2)
		ipDistanceMap[ip] = distance
	}

	if len(ipDistanceMap) == 0 {
		return "", fmt.Errorf("can not get any ip distance")
	}

	var nearestIP string
	minDistance := math.MaxFloat64
	for ip, distance := range ipDistanceMap {
		if distance < minDistance {
			minDistance = distance
			nearestIP = ip
		}
	}

	return nearestIP, nil
}
