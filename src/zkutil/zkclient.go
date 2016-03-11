// Documentation for github.com/samuel/go-zookeeper/zk: 
// http://godoc.org/github.com/samuel/go-zookeeper/zk

package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

var (
	ZKConnPool		= make(map[string]*ZKConn)
	ZKConnPoolMu	sync.Mutex
)


type ZKConn struct {
	conn	*zk.Conn
}

// create a zk connection
func NewZKConn(connstr string) (zkConn *ZKConn, err error) {
	if connstr == "" {
		err = errors.New("connection string is empty")
		return
	}
	zkhost_strs := strings.Split(connstr, ",")
	conn, echan, err := zk.Connect(zkhost_strs, 3*time.Second)
	if err != nil {
		err = fmt.Errorf("fail to connect zk server[%s]: %s", zkhost_strs, err)
		return
	}

	for {
		select {
		case connEvent := <-echan:
			switch connEvent.State {
			case zk.StateDisconnected:
				err = errors.New(fmt.Sprintf("fail to connect zk server[%s]", zkhost_strs))
				return
			case zk.StateConnected:
				zkConn = &ZKConn{
					conn: conn,
				}
				ZKConnPoolMu.Lock()
				ZKConnPool[connstr] = zkConn
				ZKConnPoolMu.Unlock()
				return
			default:
				continue
			}
		}
	}
}

// get a zk connection from the connection pool
func GetZKInstance(connstr string) (zkConn *ZKConn, err error) {
	ZKConnPoolMu.Lock()
	zkconn, ok := ZKConnPool[connstr]
	ZKConnPoolMu.Unlock()

	if ok {
		if zkconn.conn.State() == zk.StateDisconnected {
			return NewZKConn(connstr)
		}
		zkConn = zkconn
		return
	} else {
		return NewZKConn(connstr)
	}
}

func getConnstrByConn(conn *ZKConn) (connstr string, err error) {
	ZKConnPoolMu.Lock()
	defer ZKConnPoolMu.Unlock()

	for k, v := range ZKConnPool {
		if v == conn {
			connstr = k
			return
		}
	}
	err  = errors.New("fail to get the connstr by ZKConn")
	return
}

// create a znode
func (c *ZKConn) CreateNode(path string, data []byte) (resPath string, err error) {
	if path == "" {
		return "", errors.New("invalid znode path")
	}
	// 判断连接是否断开
	if c.conn.State() == zk.StateDisconnected {
		connstr, err := getConnstrByConn(c)
		if err != nil {
			return "", err
		}
		c, err = GetZKInstance(connstr)
		if err != nil {
			return "", err
		}
		return c.CreateNode(path, data)
	}

	// znode类型设置
	flag := int32(zk.FlagEphemeral)
	acl  := zk.WorldACL(zk.PermAll)
	// 创建父zknode
	childPath := path
	paths := strings.Split(path, "/")
	parentPath := strings.Join(paths[:len(paths)-1], "/")
	exist, _, err := c.conn.Exists(parentPath)
	if err != nil {
		return "", err
	}
	if !exist {
		resPath, err = c.conn.Create(parentPath, nil, 0, acl) // 父节点需是持久节点
		if err != nil {
			return
		}
	}
	// 创建子znode
	exist, stat, err := c.conn.Exists(childPath)
	if err != nil {
		return
	}
	if !exist {
		resPath, err = c.conn.Create(childPath, data, flag, acl)
		if err != nil {
			return
		}
	} else {
		_, err = c.conn.Set(childPath, data, stat.Version)
		if err != nil {
			return
		}
		resPath = childPath
	}
	return
}

// set the data of a given znode
func (c *ZKConn) SetNode(path string, data []byte) (err error) {
	if path == "" {
		return errors.New("invalid znode path")
	}
	if c.conn.State() == zk.StateDisconnected {
		connstr, err := getConnstrByConn(c)
		if err != nil {
			return err
		}
		c, err = GetZKInstance(connstr)
		if err != nil {
			return err
		}
		return c.SetNode(path, data)
	}

	exist, stat, err := c.conn.Exists(path)
	if err != nil {
		return
	}
	if !exist {
		return errors.New(fmt.Sprintf("znode[%s] doesn't exist", path))
	}
	_, err = c.conn.Set(path, data, stat.Version)
	if err != nil {
		return
	}
	return
}

// get the data of a given znode
func (c *ZKConn) GetNode(path string) (data []byte, err error) {
	if path == "" {
		return nil, errors.New("invalid znode path")
	}
	if c.conn.State() == zk.StateDisconnected {
		connstr, err := getConnstrByConn(c)
		if err != nil {
			return nil, err
		}
		c, err = GetZKInstance(connstr)
		if err != nil {
			return nil, err
		}
		return c.GetNode(path)
	}
	data, _, err = c.conn.Get(path)
	return
}

// delete a znode
func (c *ZKConn) DeleteNode(path string) (err error) {
	if path == "" {
		return errors.New("invalid znode path")
	}
	if c.conn.State() == zk.StateDisconnected {
		connstr, err := getConnstrByConn(c)
		if err != nil {
			return err
		}
		c, err = GetZKInstance(connstr)
		if err != nil {
			return err
		}
		return c.DeleteNode(path)
	}
	exist, stat, err := c.conn.Exists(path)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New(fmt.Sprintf("path[%s] doesn't exist", path))
	}
	return c.conn.Delete(path, stat.Version)
}

// list the children of a given znode
func (c *ZKConn) ListChildren(path string) (children []string, err error) {
	if path == "" {
		return nil, errors.New("invalid znode path")
	}
	if c.conn.State() == zk.StateDisconnected {
		connstr, err := getConnstrByConn(c)
		if err != nil {
			return nil, err
		}
		c, err = GetZKInstance(connstr)
		if err != nil {
			return nil, err
		}
		return c.ListChildren(path)
	}
	children, _, err = c.conn.Children(path)
	return
}
