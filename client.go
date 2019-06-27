package aliecs

import (
	ali "github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

type Client struct {
	region RegionId
	ali    *ali.Client
}

func NewClient(config *Cfg) (*Client, error) {
	c, err := ali.NewClientWithAccessKey(string(config.Derived.Region), config.AccessKeyId, config.AccessKeySecret)
	if err != nil {
		return nil, err
	}
	return &Client{region: config.Derived.Region, ali: c}, nil
}
