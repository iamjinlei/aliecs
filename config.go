package aliecs

import (
	"errors"
	"os"
)

var (
	ErrBadAccessKeyId     = errors.New("bad access key id")
	ErrBadAccessKeySecret = errors.New("bad access key secret")
	ErrBadRootPwd         = errors.New("bad root pasword")
	ErrNoMatchingRegion   = errors.New("no matching region found for zone")
)

type Derived struct {
	Region RegionId
}

type Cfg struct {
	DryRun bool

	AccessKeyId     string
	AccessKeySecret string
	KeyPairName     string
	RootPwd         string

	Zone                    ZoneId
	InstanceType            InstanceType
	Image                   ImageId
	InstanceChargeType      InstanceChargeType
	InternetChargeType      InternetChargeType
	InternetMaxBandwidthIn  int
	InternetMaxBandwidthOut int
	SystemDiskCategory      SystemDiskCategory
	SystemDiskSize          int

	InitCmds []string

	Derived Derived
}

func NewConfig() (*Cfg, error) {
	c := &Cfg{
		DryRun:          false,
		AccessKeyId:     os.Getenv("ECS_ACCESS_KEY_ID"),
		AccessKeySecret: os.Getenv("ECS_ACCESS_KEY_SECRET"),
		KeyPairName:     os.Getenv("ECS_KEY_PAIR_NAME"),
		RootPwd:         os.Getenv("ECS_ROOT_PWD"),
		Zone:            ZoneHkB,
		InstanceType:    T5c1m1,
		//Image:           CentOsV706,
		Image:                   UbuntuV1604,
		InstanceChargeType:      PostPaid,
		InternetChargeType:      PayByTraffic,
		InternetMaxBandwidthIn:  5,
		InternetMaxBandwidthOut: 5,
		SystemDiskCategory:      CloudSsd,
		SystemDiskSize:          20,

		InitCmds: []string{
			InstallUnixDev(),
		},
	}

	if len(c.AccessKeyId) == 0 {
		return nil, ErrBadAccessKeyId
	}
	if len(c.AccessKeySecret) == 0 {
		return nil, ErrBadAccessKeySecret
	}
	if len(c.RootPwd) == 0 {
		return nil, ErrBadRootPwd
	}

	region, found := ZoneToRegion[c.Zone]
	if !found {
		return nil, ErrNoMatchingRegion
	}

	c.Derived.Region = region

	return c, nil
}
