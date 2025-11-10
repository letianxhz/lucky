package rpcCenter

import (
	ccode "github.com/cherry-game/cherry/code"
	cfacade "github.com/cherry-game/cherry/faca
	clog "github.com/cherry-game/cherry/logger"
	"lucky/server/gen/msg"
)

// route = 节点类型.节点handler.remote函数

const (
	centerType = "center"
)

const (
	opsActor     = ".ops"
	accountActor = ".account"
	uuidActor    = ".uuid"
)

const (
	ping               = "ping"
	registerDevAccount = "registerDevAccount"
	getDevAccount      = "getDevAccount"
	getUID             = "getUID"
	allocateUUID       = "allocateUUID"
)

const (
	sourcePath = ".system"
)

// Ping 访问center节点，确认center已启动
func Ping(app cfacade.IApplication) bool {
	nodeID := GetCenterNodeID(app)
	if nodeID == "" {
		return false
	}

	rsp := &msg.Bool{}
	targetPath := nodeID + opsActor
	errCode := app.ActorSystem().CallWait(sourcePath, targetPath, ping, nil, rsp)
	if ccode.IsFail(errCode) {
		return false
	}

	return rsp.Value
}

// RegisterDevAccount 注册帐号
func RegisterDevAccount(app cfacade.IApplication, accountName, password, ip string) int32 {
	req := &msg.DevRegister{
		AccountName: accountName,
		Password:    password,
		Ip:          ip,
	}

	targetPath := GetTargetPath(app, accountActor)
	rsp := &msg.Int32{}
	errCode := app.ActorSystem().CallWait(sourcePath, targetPath, registerDevAccount, req, rsp)
	if ccode.IsFail(errCode) {
		clog.Warnf("[RegisterDevAccount] accountName = %s, errCode = %v", accountName, errCode)
		return errCode
	}

	return rsp.Value
}

// GetDevAccount 获取帐号信息
func GetDevAccount(app cfacade.IApplication, accountName, password string) int64 {
	req := &msg.DevRegister{
		AccountName: accountName,
		Password:    password,
	}

	targetPath := GetTargetPath(app, accountActor)
	rsp := &msg.Int64{}
	errCode := app.ActorSystem().CallWait(sourcePath, targetPath, getDevAccount, req, rsp)
	if ccode.IsFail(errCode) {
		clog.Warnf("[GetDevAccount] accountName = %s, errCode = %v", accountName, errCode)
		return 0
	}

	return rsp.Value
}

// GetUID 获取帐号UID
func GetUID(app cfacade.IApplication, sdkId, pid int32, openId string) (cfacade.UID, int32) {
	req := &msg.User{
		SdkId:  sdkId,
		Pid:    pid,
		OpenId: openId,
	}

	targetPath := GetTargetPath(app, accountActor)
	rsp := &msg.Int64{}
	errCode := app.ActorSystem().CallWait(sourcePath, targetPath, getUID, req, rsp)
	if ccode.IsFail(errCode) {
		clog.Warnf("[GetUID] errCode = %v", errCode)
		return 0, errCode
	}

	return rsp.Value, ccode.OK
}

// AllocateUUID 从 center 分配 UUID 范围（每次 1024 个）
func AllocateUUID(app cfacade.IApplication, name string) (*msg.UuidRange, int32) {
	req := &msg.String{
		Value: name,
	}

	targetPath := GetTargetPath(app, uuidActor)
	rsp := &msg.UuidRange{}
	errCode := app.ActorSystem().CallWait(sourcePath, targetPath, allocateUUID, req, rsp)
	if ccode.IsFail(errCode) {
		clog.Warnf("[AllocateUUID] name = %s, errCode = %v", name, errCode)
		return nil, errCode
	}

	return rsp, ccode.OK
}

func GetCenterNodeID(app cfacade.IApplication) string {
	list := app.Discovery().ListByType(centerType)
	if len(list) > 0 {
		return list[0].GetNodeID()
	}
	return ""
}

func GetTargetPath(app cfacade.IApplication, actorID string) string {
	nodeID := GetCenterNodeID(app)
	return nodeID + actorID
}
