package aliyun

type InstanceChargeType string

const (
	PrePaid  InstanceChargeType = "PrePaid"
	PostPaid InstanceChargeType = "PostPaid"
)

type RegionId string

const (
	RegionHz RegionId = "cn-hangzhou"
	RegionHk RegionId = "cn-hongkong"
	RegionSg RegionId = "ap-southeast-1"
)

type ZoneId string

const (
	ZoneHzB ZoneId = "cn-hangzhou-b"
	ZoneHkB ZoneId = "cn-hongkong-b"
	ZoneHkC ZoneId = "cn-hongkong-c"
	ZoneSgA ZoneId = "ap-southeast-1c"
)

var (
	ZoneToRegion = map[ZoneId]RegionId{
		ZoneHzB: RegionHz,
		ZoneHkB: RegionHk,
		ZoneHkC: RegionHk,
		ZoneSgA: RegionSg,
	}
	RegionToBr = map[RegionId]string{
		RegionHz: "hz",
		RegionHk: "hk",
		RegionSg: "sg",
	}
)

type ImageId string

const (
	CentOsV706  ImageId = "centos_7_06_64_20G_alibase_20190218.vhd"
	UbuntuV1604 ImageId = "ubuntu_16_04_64_20G_alibase_20190513.vhd"
)

type InstanceType string

const (
	T5c1m1 InstanceType = "ecs.t5-lc1m1.small" // 1Core-1GB 0.09 + 0.01/hr
	T5c1m2 InstanceType = "ecs.t5-lc1m2.small" // 1Core-2GB 0.14 + 0.01/hr
	T5c2m2 InstanceType = "ecs.t5-c1m1.large"  // 2Core-2GB 0.22 + 0.01/hr
	T5c2m4 InstanceType = "ecs.t5-lc1m2.large" // 2Core-4GB 0.31 + 0.01/hr
	T5c4m4 InstanceType = "ecs.t5-c1m1.xlarge" // 4Core-4GB 0.44 + 0.01/hr
	T5c4m8 InstanceType = "ecs.t5-c1m2.xlarge" // 4Core-8GB 0.61 + 0.01/hr
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
