package aliyun

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

func (c *EcsClient) DescribeZones(region RegionId, instanceChargeType InstanceChargeType) {
	req := ecs.CreateDescribeZonesRequest()
	req.RegionId = string(region)
	req.InstanceChargeType = string(instanceChargeType)

	response, err := c.ecs.DescribeZones(req)
	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Printf("response is %#v\n", response)
}
