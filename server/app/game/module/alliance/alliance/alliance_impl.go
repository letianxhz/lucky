package alliance

import (
	"lucky/server/pkg/di"

	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
)

// AllianceModule 联盟模块实现
// 类似 Java 项目的联盟模块，负责联盟相关的业务逻辑
type AllianceModule struct {
	// 可以注入依赖，如缓存、数据库等
	// cache alliance.IAllianceCache `di:"auto"`
	// db    *db.DB                   `di:"auto"`
}

// init 初始化联盟模块并注册到 di 容器
// 参考 claim ioc 的实现，简化注册逻辑
func init() {
	var v = &AllianceModule{}
	di.Register(v)
}

// CreateAlliance 创建联盟
func (m *AllianceModule) CreateAlliance(session *cproto.Session, req interface{}) (interface{}, error) {
	// TODO: 实现创建联盟的业务逻辑
	clog.Debugf("[AllianceModule] CreateAlliance called. uid=%d", session.Uid)
	return nil, nil
}

// JoinAlliance 加入联盟
func (m *AllianceModule) JoinAlliance(session *cproto.Session, req interface{}) (interface{}, error) {
	// TODO: 实现加入联盟的业务逻辑
	clog.Debugf("[AllianceModule] JoinAlliance called. uid=%d", session.Uid)
	return nil, nil
}

// LeaveAlliance 离开联盟
func (m *AllianceModule) LeaveAlliance(session *cproto.Session, req interface{}) error {
	// TODO: 实现离开联盟的业务逻辑
	clog.Debugf("[AllianceModule] LeaveAlliance called. uid=%d", session.Uid)
	return nil
}

// GetAllianceInfo 获取联盟信息
func (m *AllianceModule) GetAllianceInfo(session *cproto.Session, req interface{}) (interface{}, error) {
	// TODO: 实现获取联盟信息的业务逻辑
	clog.Debugf("[AllianceModule] GetAllianceInfo called. uid=%d", session.Uid)
	return nil, nil
}
