package geo

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gnasnik/titan-workerd-api/config"
	"github.com/gnasnik/titan-workerd-api/core/dao"
	"github.com/gnasnik/titan-workerd-api/core/generated/model"
	"github.com/gnasnik/titan-workerd-api/pkg/iptool"

	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("geo")

func GetIpLocation(ctx context.Context, ip string, languages ...model.Language) (*model.Location, error) {
	var lang model.Language

	if len(languages) == 0 {
		lang = model.LanguageEN
	} else {
		lang = languages[0]
	}

	// get info from databases
	var locationDb model.Location
	err := dao.GetLocationInfoByIp(ctx, ip, &locationDb, lang)
	if err == nil {
		return &locationDb, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		log.Errorf("get location by ip: %v", err)
		return nil, err
	}

	// get location from ip data cloud api
	loc, err := iptool.IPDataCloudGetLocation(ctx, config.Cfg.IpDataCloud.Url, ip, config.Cfg.IpDataCloud.Key, string(lang))
	if err != nil {
		log.Errorf("ip data cloud get location, ip: %s : %v", ip, err)
		return nil, err
	}

	if err := dao.UpsertLocationInfo(ctx, loc, lang); err != nil {
		log.Errorf("add location: %v", err)
		return nil, err
	}

	return loc, nil
}
