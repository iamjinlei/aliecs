package aliyun

import (
	"errors"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

type EcsClient struct {
	region RegionId
	ecs    *ecs.Client
}

func NewEcsClient(config *EcsCfg) (*EcsClient, error) {
	c, err := ecs.NewClientWithAccessKey(string(config.Derived.Region), config.AccessKeyId, config.AccessKeySecret)
	if err != nil {
		return nil, err
	}
	return &EcsClient{region: config.Derived.Region, ecs: c}, nil
}

const (
	vpcCidrBlock     = "172.16.0.0/12"
	vSwitchCidrBlock = "172.16.0.0/24"
)

var (
	ErrInstanceNotAvailable = errors.New("instance not available")
	ErrVpcCreation          = errors.New("unknown vpc creation error")
	ErrVSwitchCreation      = errors.New("unknown vswitch creation error")
)

func (c *EcsClient) describeVpcs(region RegionId) ([]ecs.Vpc, error) {
	req := ecs.CreateDescribeVpcsRequest()
	req.RegionId = string(region)
	resp, err := c.ecs.DescribeVpcs(req)
	if err != nil {
		return nil, err
	}
	return resp.Vpcs.Vpc, err
}

func (c *EcsClient) createVpc(region RegionId) (string, error) {
	req := ecs.CreateCreateVpcRequest()
	req.RegionId = string(region)
	req.CidrBlock = vpcCidrBlock
	resp, err := c.ecs.CreateVpc(req)
	if err != nil {
		return "", err
	}

	return resp.VpcId, nil
}

func (c *EcsClient) deleteVpc(region RegionId, vpcId string) error {
	req := ecs.CreateDeleteVpcRequest()
	req.RegionId = string(region)
	req.VpcId = vpcId
	_, err := c.ecs.DeleteVpc(req)
	return err
}

func (c *EcsClient) ensureVpc(region RegionId) (string, error) {
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		vpcs, err := c.describeVpcs(region)
		if err != nil {
			return "", err
		}

		hasPending := false
		for _, v := range vpcs {
			if v.Status == "Pending" {
				hasPending = true
			} else if v.Status == "Available" {
				return v.VpcId, nil
			}
		}

		if !hasPending {
			break
		}
	}

	return "", nil
}

func (c *EcsClient) describeVSwitches(region RegionId) ([]ecs.VSwitch, error) {
	req := ecs.CreateDescribeVSwitchesRequest()
	req.RegionId = string(region)
	resp, err := c.ecs.DescribeVSwitches(req)
	if err != nil {
		return nil, err
	}
	return resp.VSwitches.VSwitch, nil
}

func (c *EcsClient) createVSwitch(region RegionId, zone ZoneId, vpcId string) (string, error) {
	req := ecs.CreateCreateVSwitchRequest()
	req.CidrBlock = vSwitchCidrBlock
	req.VpcId = vpcId
	req.ZoneId = string(zone)
	req.RegionId = string(region)
	resp, err := c.ecs.CreateVSwitch(req)
	if err != nil {
		return "", err
	}

	return resp.VSwitchId, nil
}

func (c *EcsClient) deleteVSwitch(vSwitchId string) error {
	req := ecs.CreateDeleteVSwitchRequest()
	req.VSwitchId = vSwitchId
	_, err := c.ecs.DeleteVSwitch(req)
	return err
}

func (c *EcsClient) ensureVSwitch(region RegionId, zone ZoneId) (string, string, error) {
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		vSwitches, err := c.describeVSwitches(region)
		if err != nil {
			return "", "", err
		}

		hasPending := false
		for _, s := range vSwitches {
			if s.Status == "Pending" {
				hasPending = true
			} else if s.Status == "Available" && s.ZoneId == string(zone) {
				return s.VpcId, s.VSwitchId, nil
			}
		}

		if !hasPending {
			break
		}
	}

	return "", "", nil
}

func (c *EcsClient) ensureNetwork(region RegionId, zone ZoneId) (string, string, error) {
	vpcId, err := c.ensureVpc(region)
	if err != nil {
		return "", "", err
	}
	if vpcId == "" {
		if _, err := c.createVpc(region); err != nil {
			return "", "", err
		}
	}
	vpcId, err = c.ensureVpc(region)
	if err != nil {
		return "", "", err
	}
	if vpcId == "" {
		return "", "", ErrVpcCreation
	}

	vpcId2, vSwitchId, err := c.ensureVSwitch(region, zone)
	if err != nil {
		return "", "", err
	}
	if vSwitchId == "" {
		if _, err := c.createVSwitch(region, zone, vpcId); err != nil {
			return "", "", err
		}
	}
	vpcId2, vSwitchId, err = c.ensureVSwitch(region, zone)
	if err != nil {
		return "", "", err
	}
	if vSwitchId == "" || vpcId2 != vpcId {
		return "", "", ErrVSwitchCreation
	}

	return vpcId, vSwitchId, nil
}

func (c *EcsClient) CreateInstance(config *EcsCfg, name string) (string, error) {
	_, vSwitchId, err := c.ensureNetwork(config.Derived.Region, config.Zone)
	if err != nil {
		return "", err
	}

	req := ecs.CreateCreateInstanceRequest()

	// https://help.aliyun.com/document_detail/25499.html
	req.ZoneId = string(config.Zone)

	req.InstanceType = string(config.InstanceType)
	req.InstanceChargeType = string(config.InstanceChargeType)
	req.InstanceName = name
	req.HostName = name
	req.Password = config.RootPwd

	req.ImageId = string(config.Image)

	req.KeyPairName = config.KeyPairName
	req.InternetChargeType = string(config.InternetChargeType)
	req.InternetMaxBandwidthIn = requests.NewInteger(config.InternetMaxBandwidthIn)
	req.InternetMaxBandwidthOut = requests.NewInteger(config.InternetMaxBandwidthOut)
	req.VSwitchId = vSwitchId
	req.SystemDiskCategory = string(config.SystemDiskCategory)
	req.SystemDiskSize = requests.NewInteger(config.SystemDiskSize)

	req.CreditSpecification = "Unlimited"
	req.DryRun = requests.NewBoolean(config.DryRun)

	resp, err := c.ecs.CreateInstance(req)
	if err != nil {
		return "", err
	}

	return resp.InstanceId, nil
}

func (c *EcsClient) BindPublicIp(instanceId string) (string, error) {
	if false {
		return "", nil
	}
	req := ecs.CreateAllocatePublicIpAddressRequest()
	req.InstanceId = instanceId

	resp, err := c.ecs.AllocatePublicIpAddress(req)
	if err != nil {
		return "", err
	}
	return resp.IpAddress, nil
}

func (c *EcsClient) StartInstance(instanceId string) error {
	req := ecs.CreateStartInstanceRequest()
	req.InstanceId = instanceId

	_, err := c.ecs.StartInstance(req)
	return err
}

func (c *EcsClient) RebootInstance(instanceId string) error {
	req := ecs.CreateRebootInstanceRequest()
	req.InstanceId = instanceId

	_, err := c.ecs.RebootInstance(req)
	return err
}

func (c *EcsClient) StopInstance(instanceId string) error {
	req := ecs.CreateStopInstanceRequest()
	req.InstanceId = instanceId
	req.ForceStop = requests.NewBoolean(true)

	_, err := c.ecs.StopInstance(req)
	return err
}

func (c *EcsClient) DeleteInstance(region RegionId, instanceId string) error {
	req := ecs.CreateDeleteInstanceRequest()
	req.InstanceId = instanceId

	_, err := c.ecs.DeleteInstance(req)
	return err
}

func (c *EcsClient) DescribeInstances(region RegionId, ip string) ([]ecs.Instance, error) {
	req := ecs.CreateDescribeInstancesRequest()
	req.RegionId = string(region)

	resp, err := c.ecs.DescribeInstances(req)
	if err != nil {
		return nil, err
	}

	if ip == "" {
		return resp.Instances.Instance, nil
	}

	for _, ins := range resp.Instances.Instance {
		for _, ipAddr := range ins.PublicIpAddress.IpAddress {
			if ipAddr == ip {
				return []ecs.Instance{ins}, nil
			}
		}
	}
	return []ecs.Instance{}, nil
}
