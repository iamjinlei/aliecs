package ecs

import (
	ali "github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

type Client struct {
	region RegionId
	ali    *ali.Client
}

func NewClient(config *Cfg) (*Client, error) {
	c, err := ali.NewClientWithAccessKey(string(config.Region), config.AccessKeyId, config.AccessKeySecret)
	if err != nil {
		return nil, err
	}
	return &Client{region: config.Region, ali: c}, nil
}
