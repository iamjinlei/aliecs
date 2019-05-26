package ecs

type InstanceChargeType string

const (
	PrePaid  InstanceChargeType = "PrePaid"
	PostPaid InstanceChargeType = "PostPaid"
)

type RegionId string

const (
	RegionHk RegionId = "cn-hongkong"
)

type ZoneId string

const (
	ZoneHkB ZoneId = "cn-hongkong-b"
	ZoneHkC ZoneId = "cn-hongkong-c"
)

type ImageId string

const (
	CentOs ImageId = "centos_7_06_64_20G_alibase_20190218.vhd"
)

type InstanceType string

const (
	T5s InstanceType = "ecs.t5-lc1m2.small"
)

type InstanceStatus string

const (
	Running  InstanceStatus = "Running"
	Starting InstanceStatus = "Starting"
	Stopping InstanceStatus = "Stopping"
	Stopped  InstanceStatus = "Stopped"
)

type InternetChargeType string

const (
	PayByTraffic InternetChargeType = "PayByTraffic"
)

type VSwitchId string

const (
	HkCVs VSwitchId = "vsw-j6ch3rbwekgl875d21gyt" // c
	HkBVs VSwitchId = "vsw-j6cum1alrpi22zdrisxs6" // b
)

type SystemDiskCategory string

const (
	CloudSsd SystemDiskCategory = "cloud_ssd"
)
