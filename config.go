package ecs

import (
	"errors"
	"os"
)

var (
	ErrBadAccessKeyId     = errors.New("bad access key id")
	ErrBadAccessKeySecret = errors.New("bad access key secret")
	ErrBadRootPwd         = errors.New("bad root pasword")
)

type Cfg struct {
	DryRun bool

	AccessKeyId     string
	AccessKeySecret string
	KeyPairName     string
	RootPwd         string

	Region                  RegionId
	Zone                    ZoneId
	InstanceType            InstanceType
	Image                   ImageId
	InstanceChargeType      InstanceChargeType
	InternetChargeType      InternetChargeType
	InternetMaxBandwidthIn  int
	InternetMaxBandwidthOut int
	VSwitch                 string
	SystemDiskCategory      SystemDiskCategory
	SystemDiskSize          int

	InitCmds []string
}

func NewConfig() (*Cfg, error) {
	c := &Cfg{
		DryRun:                  false,
		AccessKeyId:             os.Getenv("ECS_ACCESS_KEY_ID"),
		AccessKeySecret:         os.Getenv("ECS_ACCESS_KEY_SECRET"),
		KeyPairName:             os.Getenv("ECS_KEY_PAIR_NAME"),
		RootPwd:                 os.Getenv("ECS_ROOT_PWD"),
		Region:                  RegionHk,
		Zone:                    ZoneHkB,
		InstanceType:            T5s,
		Image:                   CentOs,
		InstanceChargeType:      PostPaid,
		InternetChargeType:      PayByTraffic,
		InternetMaxBandwidthIn:  5,
		InternetMaxBandwidthOut: 5,
		//VSwitch:                 "vsw-j6ch3rbwekgl875d21gyt",
		VSwitch:            "vsw-j6cum1alrpi22zdrisxs6",
		SystemDiskCategory: CloudSsd,
		SystemDiskSize:     20,

		InitCmds: []string{
			"pushd ~/ && curl -LSso setup.sh https://raw.githubusercontent.com/iamjinlei/env/master/centos_dev.sh && bash setup.sh && rm -rf setup.sh && popd",
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

	return c, nil
}
