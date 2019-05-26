package ecs

import (
	"fmt"

	ali "github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

func (c *Client) DescribeZones(region RegionId, instanceChargeType InstanceChargeType) {
	req := ali.CreateDescribeZonesRequest()
	req.RegionId = string(region)
	req.InstanceChargeType = string(instanceChargeType)

	response, err := c.ali.DescribeZones(req)
	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Printf("response is %#v\n", response)
}
