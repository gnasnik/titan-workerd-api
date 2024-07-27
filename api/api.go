package api

import (
	"context"
	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/client"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/Filecoin-Titan/titan/lib/etcdcli"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-workerd-api/config"
	"github.com/gnasnik/titan-workerd-api/core/errors"
	logging "github.com/ipfs/go-log"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
)

var log = logging.Logger("api")

var GlobalServer *Server

type Server struct {
	cfg        config.Config
	router     *gin.Engine
	etcdClient *etcdClient
}

type Scheduler struct {
	Url    string
	AreaId string
	Api    api.Scheduler
	Closer func()
}

type etcdClient struct {
	cli *etcdcli.Client

	mu         sync.Mutex
	schedulers []*Scheduler
}

func newEtcdClient(user, password string, addresses []string) (*etcdClient, error) {
	os.Setenv("ETCD_USERNAME", user)
	os.Setenv("ETCD_PASSWORD", password)

	etcd, err := etcdcli.New(addresses)
	if err != nil {
		return nil, err
	}

	client := &etcdClient{
		cli:        etcd,
		schedulers: make([]*Scheduler, 0),
	}

	return client, nil
}

func (ec *etcdClient) updateSchedulers(schedulers []*Scheduler) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.schedulers = schedulers
}

func (ec *etcdClient) getSchedulers() []*Scheduler {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	return ec.schedulers
}

func (ec *etcdClient) loadSchedulerConfigs() (map[string][]*types.SchedulerCfg, error) {
	resp, err := ec.cli.GetServers(types.NodeScheduler.String())
	if err != nil {
		return nil, err
	}

	schedulerConfigs := make(map[string][]*types.SchedulerCfg)

	for _, kv := range resp.Kvs {
		var configScheduler *types.SchedulerCfg
		err := etcdcli.SCUnmarshal(kv.Value, &configScheduler)
		if err != nil {
			return nil, err
		}
		configs, ok := schedulerConfigs[configScheduler.AreaID]
		if !ok {
			configs = make([]*types.SchedulerCfg, 0)
		}
		configs = append(configs, configScheduler)

		schedulerConfigs[configScheduler.AreaID] = configs
	}

	return schedulerConfigs, nil
}

func (s *Server) watchEtcdSchedulerConfig(ctx context.Context) {
	watchChan := s.etcdClient.cli.WatchServers(context.Background(), types.NodeScheduler.String())
	for {
		resp, ok := <-watchChan
		if !ok {
			log.Errorf("close watch chan")
			return
		}

		for _, event := range resp.Events {
			switch event.Type {
			case mvccpb.DELETE, mvccpb.PUT:
				log.Infof("Etcd Scheduler config changed")
				schedulers, err := s.fetchSchedulersFromEtcd(ctx)
				if err != nil {
					log.Errorf("FetchSchedulersFromEtcd: %v", err)
					continue
				}

				s.etcdClient.updateSchedulers(schedulers)

				log.Infof("Updated Scheduler from etcd")
			}
		}
	}
}

func (s *Server) GetSchedulers() []*Scheduler {
	return s.etcdClient.getSchedulers()
}

func (s *Server) fetchSchedulersFromEtcd(ctx context.Context) ([]*Scheduler, error) {
	schedulerConfigs, err := s.etcdClient.loadSchedulerConfigs()
	if err != nil {
		log.Errorf("load scheduer from etcd: %v", err)
		return nil, err
	}

	var schedulers []*Scheduler

	for key, schedulerURLs := range schedulerConfigs {
		for _, SchedulerCfg := range schedulerURLs {
			// https protocol still in test, we use http for now.
			schedulerURL := strings.Replace(SchedulerCfg.SchedulerURL, "https", "http", 1)
			headers := http.Header{}
			headers.Add("Authorization", "Bearer "+SchedulerCfg.AccessToken)
			clientInit, closeScheduler, err := client.NewScheduler(ctx, schedulerURL, headers)
			if err != nil {
				log.Errorf("create Scheduler rpc client: %v", err)
			}
			schedulers = append(schedulers, &Scheduler{
				Url:    schedulerURL,
				Api:    clientInit,
				AreaId: key,
				Closer: closeScheduler,
			})
		}
	}

	log.Infof("fetch %d schedulers from Etcd", len(schedulers))

	return schedulers, nil
}

func NewServer(cfg config.Config) (*Server, error) {
	gin.SetMode(cfg.Mode)
	router := gin.Default()

	// logging request body
	router.Use(RequestLoggerMiddleware())

	RegisterRouter(router, cfg)

	ec, err := newEtcdClient(cfg.EtcdUser, cfg.EtcdPassword, cfg.EtcdAddresses)
	if err != nil {
		log.Errorf("New etcdClient Failed: %v", err)
		return nil, err
	}

	s := &Server{
		cfg:        cfg,
		router:     router,
		etcdClient: ec,
	}

	schedulers, err := s.fetchSchedulersFromEtcd(context.Background())
	if err != nil {
		return nil, err
	}

	s.etcdClient.updateSchedulers(schedulers)

	go s.watchEtcdSchedulerConfig(context.Background())

	GlobalServer = s

	return s, nil
}

func (s *Server) Run() {
	err := s.router.Run(s.cfg.ApiListen)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) Close() {}

func GetSchedulerByAreaId(areaId string) (*Scheduler, error) {
	schedulers := GlobalServer.GetSchedulers()

	for _, scheduler := range schedulers {
		if scheduler.AreaId == areaId {
			return scheduler, nil
		}
	}

	return nil, errors.ErrNoAvailableScheduler
}

func GetMaybeBestScheduler(areaId string) (*Scheduler, error) {
	schedulers := GlobalServer.GetSchedulers()

	var areaIds []string
	for _, scheduler := range schedulers {
		areaIds = append(areaIds, scheduler.AreaId)
	}

	maybeBest := findBestMatch(areaIds, areaId)

	return GetSchedulerByAreaId(maybeBest)
}

// findBestMatch 用于找出数组中与目标字符串匹配最多的字符串
func findBestMatch(stringsArr []string, target string) string {
	var bestMatch string
	maxMatchLength := 0

	for _, str := range stringsArr {
		matchLength := 0

		// 将数组中的字符串和目标字符串都按照 '-' 分割
		strParts := strings.Split(str, "-")
		targetParts := strings.Split(target, "-")

		// 比较分割后的每部分，计算匹配的长度
		for i := 0; i < len(strParts) && i < len(targetParts); i++ {
			if strParts[i] == targetParts[i] {
				matchLength++
			} else {
				break
			}
		}

		// 如果当前字符串的匹配长度更长，更新最佳匹配
		if matchLength > maxMatchLength {
			maxMatchLength = matchLength
			bestMatch = str
		}
	}

	return bestMatch
}

func GetRandomSchedulerAPI() (*Scheduler, error) {
	schedulers := GlobalServer.GetSchedulers()

	if len(schedulers) == 0 {
		return nil, errors.ErrNoAvailableScheduler
	}

	//  for test
	//whitelist := []string{"Asia-Hong-Kong-2", "Asia-Hong-Kong-3", "Asia-Hong-Kong-15"}
	whitelist := []string{"Asia-China-Guangdong-Shenzhen"}
	for _, s := range schedulers {
		if s.AreaId == whitelist[rand.Intn(len(whitelist))] {
			return s, nil
		}
	}

	return schedulers[rand.Intn(len(schedulers))], nil
}
