package main

import (
	"go-zk/zk"
	"errors"
	"time"
	"fmt"
)

type ZookeeperConfig struct {
	Servers    []string
	RootPath   string
	MasterPath string
}

type ElectionManager struct {
	ZKClientConn *zk.Conn
	ZKConfig     *ZookeeperConfig
	IsMasterQ    chan bool
}

func NewElectionManager(zkConfig *ZookeeperConfig, isMasterQ chan bool) *ElectionManager {
	electionManager := &ElectionManager{
		nil,
		zkConfig,
		isMasterQ,
	}
	electionManager.initConnection()
	return electionManager
}

func (electionManager *ElectionManager) Run() {
	err := electionManager.electMaster()
	if err != nil {
		fmt.Println("elect master error, ", err)
	}
	electionManager.watchMaster()
}

// check if the connection to zookeeper is successful
func (electionManager *ElectionManager) isConnected() bool {
	if electionManager.ZKClientConn == nil {
		return false
	} else if electionManager.ZKClientConn.State() != zk.StateConnected {
		return false
	}
	return true
}

// initialize the connection to zk
func (electionManager *ElectionManager) initConnection() error {
	if !electionManager.isConnected() {
		conn, connChan, err := zk.Connect(electionManager.ZKConfig.Servers, time.Second)
		if err != nil {
			return err
		}

		for {
			isConnected := false
			select {
			case connEvent := <-connChan:
				if connEvent.State == zk.StateConnected {
					isConnected = true
					fmt.Println("connect to zookeeper server success!")
				}
			case _ = <-time.After(time.Second * 3):
				return errors.New("connect to zookeeper server timeout!")
			}
			if isConnected {
				break
			}
		}
		electionManager.ZKClientConn = conn
	}
	return nil
}

// elect the master
func (electionManager *ElectionManager) electMaster() error {
	err := electionManager.initConnection()
	if err != nil {
		return err
	}
	// check if the root name node is already existing, create one if not.
	isExist, _, err := electionManager.ZKClientConn.Exists(electionManager.ZKConfig.RootPath)
	if err != nil {
		return err
	}
	if !isExist {
		path, err := electionManager.ZKClientConn.Create(electionManager.ZKConfig.RootPath, nil, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			return err
		}
		if electionManager.ZKConfig.RootPath != path {
			return errors.New("Create returned different path " + electionManager.ZKConfig.RootPath + " != " + path)
		}
	}

	// create ephemeral znode for electing
	masterPath := electionManager.ZKConfig.RootPath + electionManager.ZKConfig.MasterPath
	path, err := electionManager.ZKClientConn.Create(masterPath, nil, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err == nil {
		if path == masterPath {
			fmt.Println("elect master success!")
			electionManager.IsMasterQ <- true
		} else {
			return errors.New("Create returned different path " + masterPath + " != " + path)
		}
	} else {
		fmt.Printf("elect master failure, ", err)
		electionManager.IsMasterQ <- false
	}
	return nil
}

// watch the master znode, in case the master server will crash.
func (electionManager *ElectionManager) watchMaster() {
	for {
		children, state, childCh, err := electionManager.ZKClientConn.ChildrenW(electionManager.ZKConfig.RootPath + electionManager.ZKConfig.MasterPath)
		if err != nil {
			fmt.Println("watch children error, ", err)
		}
		fmt.Println("watch children result, ", children, state)
		select {
		case childEvent := <-childCh:
			if childEvent.Type == zk.EventNodeDeleted {
				fmt.Println("receive znode delete event, ", childEvent)
				fmt.Println("start elect new master ...")
				err = electionManager.electMaster()
				if err != nil {
					fmt.Println("elect new master error, ", err)
				}
			}
		}
	}
}
