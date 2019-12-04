package aliyun

import (
	"strconv"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/domain"
)

type DomainClient struct {
	region RegionId
	domain *domain.Client
}

func NewDomainClient(config *DomainCfg) (*DomainClient, error) {
	c, err := domain.NewClientWithAccessKey(string(config.Derived.Region), config.AccessKeyId, config.AccessKeySecret)
	if err != nil {
		return nil, err
	}
	return &DomainClient{region: config.Derived.Region, domain: c}, nil
}

func (c *DomainClient) ListDomains() ([]domain.Domain, error) {
	req := domain.CreateQueryDomainListRequest()

	req.PageNum = requests.NewInteger(0)
	req.PageSize = requests.NewInteger(10)
	req.OrderKeyType = "RegistrationDate"

	resp, err := c.domain.QueryDomainList(req)
	if err != nil {
		return nil, err
	}
	return resp.Data.Domain, nil
}

/*
 * 1：可注册；
 * 3：预登记；
 * 4：可删除预订；
 * 0：不可注册；
 * -1：异常；
 * -2：暂停注册；
 * -3：黑名单。
 */
func (c *DomainClient) CheckDomain(d string) (string, int, string, int64, error) {
	req := domain.CreateCheckDomainRequest()

	req.DomainName = d
	req.FeeCurrency = "CNY"
	req.FeeCommand = "create"
	req.FeePeriod = requests.NewInteger(1)

	resp, err := c.domain.CheckDomain(req)
	if err != nil {
		return "", -1, "", 0, err
	}

	status, err := strconv.ParseInt(resp.Avail, 10, 64)
	if err != nil {
		return "", -1, "", 0, err
	}
	return resp.DomainName, int(status), resp.Reason, resp.Price, nil
}
