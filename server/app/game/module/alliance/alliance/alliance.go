package alliance

import (
	cproto "github.com/cherry-game/cherry/net/proto"
)

// IAllianceModule 联盟模块接口
// 定义联盟相关的所有业务操作
// 注意：这里使用 interface{} 作为占位符，实际使用时应该使用具体的 protobuf 消息类型
type IAllianceModule interface {
	// CreateAlliance 创建联盟
	// req 和 response 类型应该使用具体的 protobuf 消息类型，如 *msg.CreateAllianceRequest
	CreateAlliance(session *cproto.Session, req interface{}) (interface{}, error)

	// JoinAlliance 加入联盟
	JoinAlliance(session *cproto.Session, req interface{}) (interface{}, error)

	// LeaveAlliance 离开联盟
	LeaveAlliance(session *cproto.Session, req interface{}) error

	// GetAllianceInfo 获取联盟信息
	GetAllianceInfo(session *cproto.Session, req interface{}) (interface{}, error)
}
