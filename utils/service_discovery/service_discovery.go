package serviceDiscovery

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	jsoniter "github.com/json-iterator/go"
)

const (
	keyPrefix     = "/services"
	pathHierarchy = 4
)

type ServiceDiscovery interface {

	// 注册一个服务，当服务启时触发
	RegistrationService(ctx context.Context, s ServiceInfo, expireTime time.Duration) error

	// 注销服务, 当服务关闭时触发
	DestroyService(ctx context.Context, s ServiceInfo) error

	// 根据服务运行环境和名称检索服务
	GetServicesByEnvironmentAndName(ctx context.Context, environment, name string) (Services, error)

	// 根据服务运行环境检索服务
	GetServicesByEnvironment(ctx context.Context, environment string) (Services, error)

	// 根据服务名称检索服务
	GetServicesByName(ctx context.Context, name string) (Services, error)

	// 根据服务唯一标识检索服务
	GetServiceInfoByID(ctx context.Context, environment, name, id string) (ServiceInfo, error)

	// 获取所有服务信息
	GetAllServices(ctx context.Context) (Services, error)

	// 监控服务器变化
	WatchServiceChange(environment, name string, f func(response *clientv3.WatchResponse))

	// 获取服务器数量
	GetServiceCount(ctx context.Context, environment, name string) (int, error)
}

// 服务器组信息
type Services []*ServiceInfo

// 服务器信息
type ServiceInfo struct {

	// 服务标识，在同一类型服务下唯一，推荐ip:port格式
	ID string `json:"id"`

	// 服务名称，一组相同的服务名称相同，推荐应用名称或git仓库名称
	Name string `json:"name"`

	// 服务运行环境， 推荐: testing, debug，production
	Environment string `json:"environment"`

	// 服务运行的内网ip，可选配置
	InternalIp string `json:"internal_ip,omitempty"`

	// 服务运行的公网ip， 可选配置
	PublicIp string `json:"public_ip,omitempty"`

	// 服务监听的端口, 可选配置
	Port int32 `json:"port,omitempty"`

	// 可单独访问服务的地址，推荐ip:port
	Url string `json:"url,omitempty"`

	// Reload 接口url
	ReloadUrl string `json:"reload_url"`

	// 服务当前部署的版本
	GitVersion string `json:"git_version,omitempty"`

	// 服务启动的时间
	Time time.Time `json:"time,omitempty"`

	// 服务监控，必须提供监控接口
	IsMetricsApi bool `json:"is_metrics_api,omitempty"`

	Data map[string]string `json:"data,omitempty"`
}

type discovery struct {
	cli *clientv3.Client
}

func New(cli *clientv3.Client) ServiceDiscovery {
	return &discovery{
		cli: cli,
	}
}

func (d *discovery) RegistrationService(ctx context.Context, s ServiceInfo, expireTime time.Duration) error {
	b, err := jsoniter.Marshal(s)
	if err != nil {
		return err
	}

	resp, err := d.cli.Grant(ctx, int64(expireTime.Seconds()))
	if err != nil {
		return err
	}

	res, err := d.cli.Txn(ctx).Then(
		clientv3.OpPut(fmt.Sprintf("%s/%s/%s/%s", keyPrefix, s.Environment, s.Name, s.ID), string(b), clientv3.WithLease(resp.ID)),
		clientv3.OpPut(fmt.Sprintf("%s/%s/%s", keyPrefix, s.Environment, s.Name), s.Name, clientv3.WithLease(resp.ID)),
	).Commit()

	if err != nil {
		return err
	}

	if !res.Succeeded {
		return errors.New("service registration fail")
	}

	return nil
}

func (d *discovery) DestroyService(ctx context.Context, s ServiceInfo) error {
	_, err := d.cli.Delete(ctx, fmt.Sprintf("%s/%s/%s/%s", keyPrefix, s.Environment, s.Name, s.ID))
	return err
}

func (d *discovery) GetServicesByEnvironmentAndName(ctx context.Context, environment, name string) (Services, error) {
	data, err := d.getServicesByPrefix(ctx, fmt.Sprintf("%s/%s/%s/", keyPrefix, environment, name))
	if err != nil {
		return nil, err
	}

	res := make(Services, 0, len(data))
	for _, v := range data {
		if v.Environment != environment || v.Name != name {
			continue
		}

		res = append(res, v)
	}
	return res, nil
}

func (d *discovery) GetServicesByEnvironment(ctx context.Context, environment string) (Services, error) {
	data, err := d.getServicesByPrefix(ctx, fmt.Sprintf("%s/%s/", keyPrefix, environment))
	if err != nil {
		return nil, err
	}

	res := make(Services, 0, len(data))
	for _, v := range data {
		if v.Environment != environment {
			continue
		}

		res = append(res, v)
	}
	return res, nil
}

func (d *discovery) GetServicesByName(ctx context.Context, name string) (Services, error) {
	res := make(Services, 0)

	s, err := d.getServicesByPrefix(ctx, keyPrefix)
	if err != nil {
		return res, err
	}

	for _, info := range s {
		if info.Name == name {
			res = append(res, info)
		}
	}

	return res, nil
}

func (d *discovery) GetServiceInfoByID(ctx context.Context, environment, name, id string) (ServiceInfo, error) {

	s, err := d.getServicesByPrefix(ctx, fmt.Sprintf("%s/%s/%s/%s", keyPrefix, environment, name, id))
	if err != nil {
		return ServiceInfo{}, err
	}

	if len(s) > 0 {
		return *s[0], nil
	}

	return ServiceInfo{}, errors.New("service information not found")
}

func (d *discovery) GetAllServices(ctx context.Context) (Services, error) {
	return d.getServicesByPrefix(ctx, keyPrefix)
}

func (d *discovery) getServicesByPrefix(ctx context.Context, prefix string) (Services, error) {
	res := make(Services, 0)
	data, err := d.cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return res, err
	}

	for _, kv := range data.Kvs {
		if strings.Count(string(kv.Key), "/") < pathHierarchy {
			continue
		}

		sInfo := &ServiceInfo{}
		err = jsoniter.Unmarshal(kv.Value, sInfo)
		if err != nil {
			return res, err
		}

		res = append(res, sInfo)
	}

	return res, nil
}

func (d *discovery) WatchServiceChange(environment, name string, f func(response *clientv3.WatchResponse)) {
	watchChan := d.cli.Watch(context.Background(), fmt.Sprintf("%s/%s/%s/", keyPrefix, environment, name), clientv3.WithPrefix())
	for watchResponse := range watchChan {
		f(&watchResponse)
	}
}

func (d *discovery) GetServiceCount(ctx context.Context, environment, name string) (int, error) {
	data, err := d.getServicesByPrefix(ctx, fmt.Sprintf("%s/%s/%s/", keyPrefix, environment, name))
	if err != nil {
		return 0, err
	}

	var count int
	for _, v := range data {
		if v.Environment != environment || v.Name != name {
			continue
		}

		count++
	}

	return count, nil
}
