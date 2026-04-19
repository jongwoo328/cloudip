package provider

import (
	"cloudip/util"
	"fmt"
	"net"
	"sync"
)

type CloudProvider interface {
	Initialize() error
	CheckParsedIP(parsedIP net.IP) (bool, error)
	GetName() string
}

type DataManager interface {
	EnsureDataFile() error
	GetDataURL() string
}

type BaseProvider struct {
	name        string
	v4Tree      *util.CIDRTree
	v6Tree      *util.CIDRTree
	initialized bool
	initLock    sync.Mutex
	dataManager DataManager
	loadFunc    func(*BaseProvider) error
}

func NewBaseProvider(name string, dataManager DataManager, loadFunc func(*BaseProvider) error) *BaseProvider {
	return &BaseProvider{
		name:        name,
		dataManager: dataManager,
		loadFunc:    loadFunc,
	}
}

func (bp *BaseProvider) GetName() string {
	return bp.name
}

func (bp *BaseProvider) CheckParsedIP(parsedIP net.IP) (bool, error) {
	if parsedIP == nil {
		return false, fmt.Errorf("error parsing IP: %v", parsedIP)
	}

	if parsedIP.To4() != nil {
		return bp.v4Tree.MatchParsedIP(parsedIP), nil
	}

	if parsedIP.To16() == nil {
		return false, fmt.Errorf("error parsing IP: %v", parsedIP)
	}

	return bp.v6Tree.MatchParsedIP(parsedIP), nil
}

func (bp *BaseProvider) Initialize() error {
	if bp.initialized {
		return nil
	}

	bp.initLock.Lock()
	defer bp.initLock.Unlock()

	if bp.initialized {
		return nil
	}

	bp.v4Tree = util.NewCIDRTree()
	bp.v6Tree = util.NewCIDRTree()

	err := bp.dataManager.EnsureDataFile()
	if err != nil {
		return err
	}

	if err := bp.loadFunc(bp); err != nil {
		return err
	}

	bp.initialized = true
	return nil
}

func (bp *BaseProvider) AddIPv4Range(cidr string) {
	bp.v4Tree.AddCIDR(cidr)
}

func (bp *BaseProvider) AddIPv6Range(cidr string) {
	bp.v6Tree.AddCIDR(cidr)
}

func (bp *BaseProvider) AddCIDRRange(cidr string) error {
	cidrVersion, err := util.GetCIDRVersion(cidr)
	if err != nil {
		return err
	}

	if cidrVersion == util.IPv4 {
		bp.AddIPv4Range(cidr)
	} else {
		bp.AddIPv6Range(cidr)
	}
	return nil
}
