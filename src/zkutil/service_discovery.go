package main

import (
	"fmt"
	"errors"
	"math/rand"
	"github.com/golang/protobuf/proto"
)

const (
	zkservers = "172.16.130.1:2181,172.16.130.2:2181,172.16.130.3:2181,172.16.181.1:2181,172.16.181.2:2181"
)

type ServiceNode struct {
	Ip		*string
	Port	*uint32
}

func (s *ServiceNode) Reset()			{ *s = ServiceNode{} }
func (s *ServiceNode) String() string	{ return proto.CompactTextString(s)}
func (*ServiceNode) ProtoMessage()		{}

func (s *ServiceNode) GetIp() string {
	if s != nil && s.Ip != nil {
		return *s.Ip
	}
	return ""
}

func (s *ServiceNode) GetPort() uint32 {
	if s != nil && s.Port != nil {
		return *s.Port
	}
	return 0
}


func Register(zkpath string, ip string, port uint32) (err error) {
	service_node := &ServiceNode{
		Ip:		proto.String(ip),
		Port:	proto.Uint32(port),
	}

	// 将IP和Port信息编码为二进制流
	buffer_service_node, err := proto.Marshal(service_node)
	if err != nil {
		return errors.New("serilize service node fail")
	}

	// 注册
	zkconn, err := GetZKInstance(zkservers)
	if err != nil {
		return errors.New("fail to connect to zk hosts")
	}
	res, err := zkconn.CreateNode(zkpath, buffer_service_node)
	if err != nil {
		return errors.New("fail to register")
	} else {
		fmt.Printf("register %s successfully", res)
		return nil
	}
}

func Discovery(zkpath string) (ip string, port uint32, err error) {
	zkconn, err := GetZKInstance(zkservers)

	// 获取zkpath所有的叶子节点路径
	child_znodes, _ := zkconn.ListChildren(zkpath)
	// 随机获取一个注册实例，可用于负载均衡
	num := rand.Intn(len(child_znodes))
	real_node_path := zkpath + "/" + child_znodes[num]
	// 获取znode路径上的值
	buffer, err := zkconn.GetNode(real_node_path)
	if err != nil {
		return
	}

	// 对数据解码
	service_node := &ServiceNode{}
	err = proto.Unmarshal(buffer, service_node)
	if err != nil {
		return
	}
	return service_node.GetIp(), service_node.GetPort(), nil
}
