package ecs

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	ali "github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

func (c *Client) CreateInstance(config *Cfg, name string) (string, error) {
	req := ali.CreateCreateInstanceRequest()

	// https://help.aliyun.com/document_detail/25499.html
	req.ZoneId = string(config.Zone)

	req.InstanceType = string(config.InstanceType)
	req.InstanceChargeType = string(config.InstanceChargeType)
	req.InstanceName = name
	req.HostName = name
	req.Password = config.RootPwd

	req.ImageId = string(config.Image)

	req.KeyPairName = "ecs_key"
	req.InternetChargeType = string(config.InternetChargeType)
	req.InternetMaxBandwidthIn = requests.NewInteger(config.InternetMaxBandwidthIn)
	req.InternetMaxBandwidthOut = requests.NewInteger(config.InternetMaxBandwidthOut)
	req.VSwitchId = string(config.VSwitch)
	req.SystemDiskCategory = string(config.SystemDiskCategory)
	req.SystemDiskSize = requests.NewInteger(config.SystemDiskSize)

	req.CreditSpecification = "Unlimited"
	req.DryRun = requests.NewBoolean(config.DryRun)

	resp, err := c.ali.CreateInstance(req)
	if err != nil {
		return "", err
	}

	return resp.InstanceId, nil
}

func (c *Client) BindPublicIp(instanceId string) (string, error) {
	if false {
		return "", nil
	}
	req := ali.CreateAllocatePublicIpAddressRequest()
	req.InstanceId = instanceId

	resp, err := c.ali.AllocatePublicIpAddress(req)
	if err != nil {
		return "", err
	}
	return resp.IpAddress, nil
}

func (c *Client) StartInstance(instanceId string) error {
	req := ali.CreateStartInstanceRequest()
	req.InstanceId = instanceId

	_, err := c.ali.StartInstance(req)
	return err
}

func (c *Client) StopInstance(instanceId string) error {
	req := ali.CreateStopInstanceRequest()
	req.InstanceId = instanceId
	req.ForceStop = requests.NewBoolean(true)

	_, err := c.ali.StopInstance(req)
	return err
}

func (c *Client) DelInstance(instanceId string) error {
	req := ali.CreateDeleteInstanceRequest()
	req.InstanceId = instanceId

	_, err := c.ali.DeleteInstance(req)
	return err
}

func (c *Client) DescribeInstances(instanceNameRegex string) ([]ali.Instance, error) {
	req := ali.CreateDescribeInstancesRequest()
	req.RegionId = string(c.region)
	req.InstanceName = instanceNameRegex

	resp, err := c.ali.DescribeInstances(req)
	if err != nil {
		return nil, err
	}

	return resp.Instances.Instance, nil
}
