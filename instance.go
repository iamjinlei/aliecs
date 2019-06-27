package aliecs

import (
	"errors"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	ali "github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

const (
	vpcCidrBlock     = "172.16.0.0/12"
	vSwitchCidrBlock = "172.16.0.0/24"
)

var (
	ErrInstanceNotAvailable = errors.New("instance not available")
	ErrVpcCreation          = errors.New("unknown vpc creation error")
	ErrVSwitchCreation      = errors.New("unknown vswitch creation error")
)

func (c *Client) describeVpcs(region RegionId) ([]ali.Vpc, error) {
	req := ali.CreateDescribeVpcsRequest()
	req.RegionId = string(region)
	resp, err := c.ali.DescribeVpcs(req)
	if err != nil {
		return nil, err
	}
	return resp.Vpcs.Vpc, err
}

func (c *Client) createVpc(region RegionId) (string, error) {
	req := ali.CreateCreateVpcRequest()
	req.RegionId = string(region)
	req.CidrBlock = vpcCidrBlock
	resp, err := c.ali.CreateVpc(req)
	if err != nil {
		return "", err
	}

	return resp.VpcId, nil
}

func (c *Client) deleteVpc(region RegionId, vpcId string) error {
	req := ali.CreateDeleteVpcRequest()
	req.RegionId = string(region)
	req.VpcId = vpcId
	_, err := c.ali.DeleteVpc(req)
	return err
}

func (c *Client) ensureVpc(region RegionId) (string, error) {
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

func (c *Client) describeVSwitches(region RegionId) ([]ali.VSwitch, error) {
	req := ali.CreateDescribeVSwitchesRequest()
	req.RegionId = string(region)
	resp, err := c.ali.DescribeVSwitches(req)
	if err != nil {
		return nil, err
	}
	return resp.VSwitches.VSwitch, nil
}

func (c *Client) createVSwitch(region RegionId, zone ZoneId, vpcId string) (string, error) {
	req := ali.CreateCreateVSwitchRequest()
	req.CidrBlock = vSwitchCidrBlock
	req.VpcId = vpcId
	req.ZoneId = string(zone)
	req.RegionId = string(region)
	resp, err := c.ali.CreateVSwitch(req)
	if err != nil {
		return "", err
	}

	return resp.VSwitchId, nil
}

func (c *Client) deleteVSwitch(vSwitchId string) error {
	req := ali.CreateDeleteVSwitchRequest()
	req.VSwitchId = vSwitchId
	_, err := c.ali.DeleteVSwitch(req)
	return err
}

func (c *Client) ensureVSwitch(region RegionId, zone ZoneId) (string, string, error) {
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

func (c *Client) ensureNetwork(region RegionId, zone ZoneId) (string, string, error) {
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

func (c *Client) CreateInstance(config *Cfg, name string) (string, error) {
	_, vSwitchId, err := c.ensureNetwork(config.Derived.Region, config.Zone)
	if err != nil {
		return "", err
	}

	req := ali.CreateCreateInstanceRequest()

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

func (c *Client) DeleteInstance(region RegionId, instanceId string) error {
	req := ali.CreateDeleteInstanceRequest()
	req.InstanceId = instanceId

	_, err := c.ali.DeleteInstance(req)
	return err
}

func (c *Client) DescribeInstances(region RegionId, ip string) ([]ali.Instance, error) {
	req := ali.CreateDescribeInstancesRequest()
	req.RegionId = string(region)

	resp, err := c.ali.DescribeInstances(req)
	if err != nil {
		return nil, err
	}

	if ip == "" {
		return resp.Instances.Instance, nil
	}

	for _, ins := range resp.Instances.Instance {
		for _, ipAddr := range ins.PublicIpAddress.IpAddress {
			if ipAddr == ip {
				return []ali.Instance{ins}, nil
			}
		}
	}
	return []ali.Instance{}, nil
}
