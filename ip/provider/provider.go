package provider

import (
	"cloudip/common"
	"cloudip/util"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
)

type CloudProvider interface {
	Initialize() error
	CheckParsedIP(parsedIP net.IP) (bool, error)
	GetName() string
}

type DataManager interface {
	EnsureDataFile() error
}

type UpdatePolicySetter interface {
	SetUpdatePolicy(common.UpdatePolicy)
}

type BaseProvider struct {
	name        string
	v4Tree      *util.CIDRTree
	v6Tree      *util.CIDRTree
	initialized atomic.Bool
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

func (bp *BaseProvider) SetUpdatePolicy(policy common.UpdatePolicy) {
	if setter, ok := bp.dataManager.(UpdatePolicySetter); ok {
		setter.SetUpdatePolicy(policy)
	}
}

func (bp *BaseProvider) GetName() string {
	return bp.name
}

func (bp *BaseProvider) CheckParsedIP(parsedIP net.IP) (bool, error) {
	if !bp.initialized.Load() {
		return false, fmt.Errorf("provider %s is not initialized", bp.name)
	}

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
	if bp.initialized.Load() {
		return nil
	}

	bp.initLock.Lock()
	defer bp.initLock.Unlock()

	if bp.initialized.Load() {
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

	bp.initialized.Store(true)
	return nil
}

func (bp *BaseProvider) AddIPv4Range(cidr string) error {
	if bp == nil {
		return errors.New("provider is not initialized")
	}
	if bp.v4Tree == nil {
		return fmt.Errorf("provider %s is not initialized", bp.name)
	}
	return bp.v4Tree.AddCIDR(cidr)
}

func (bp *BaseProvider) AddIPv6Range(cidr string) error {
	if bp == nil {
		return errors.New("provider is not initialized")
	}
	if bp.v6Tree == nil {
		return fmt.Errorf("provider %s is not initialized", bp.name)
	}
	return bp.v6Tree.AddCIDR(cidr)
}

func (bp *BaseProvider) AddCIDRRange(cidr string) error {
	cidrVersion, err := util.GetCIDRVersion(cidr)
	if err != nil {
		return err
	}

	if cidrVersion == util.IPv4 {
		return bp.AddIPv4Range(cidr)
	}
	return bp.AddIPv6Range(cidr)
}
