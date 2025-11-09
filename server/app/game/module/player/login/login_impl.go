package login

import (
	"fmt"
	"lucky/server/app/game/db"
	"lucky/server/app/game/module/shared/online"
	"lucky/server/gen/msg"
	"lucky/server/pkg/code"
	"lucky/server/pkg/data"
	"lucky/server/pkg/di"
	"lucky/server/pkg/event"
	sessionKey "lucky/server/pkg/session_key"

	cstring "github.com/cherry-game/cherry/extend/string"
	clog "github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/parser/pomelo"
	cproto "github.com/cherry-game/cherry/net/proto"
)

// LoginModule 登录模块实现
type LoginModule struct {
	// 可以注入依赖，如其他模块等
}

// init 初始化登录模块并注册到 di 容器
func init() {
	var v = &LoginModule{}
	di.Register(v)
}

// SelectPlayer 查询角色列表
func (m *LoginModule) SelectPlayer(session *cproto.Session) (*msg.PlayerSelectResponse, error) {
	response := &msg.PlayerSelectResponse{}

	playerId := db.GetPlayerIdWithUID(session.Uid)
	if playerId > 0 {
		// 游戏设定单服单角色，协议设计成可返回多角色
		playerTable, found := db.GetPlayerTable(playerId)
		if found {
			playerInfo := buildPBPlayer(playerTable)
			response.List = append(response.List, &playerInfo)
		}
	}

	return response, nil
}

// CreatePlayer 创建角色
func (m *LoginModule) CreatePlayer(session *cproto.Session, req *msg.PlayerCreateRequest, actor *pomelo.ActorBase) (*msg.PlayerCreateResponse, error) {
	if req.Gender > 1 {
		return nil, fmt.Errorf("invalid gender: %d", req.Gender)
	}

	// 检查玩家昵称
	if len(req.PlayerName) < 1 {
		return nil, fmt.Errorf("player name is empty")
	}

	// 帐号是否已经在当前游戏服存在角色
	if db.GetPlayerIdWithUID(session.Uid) > 0 {
		return nil, fmt.Errorf("player already exists")
	}

	// 获取创角初始化配置
	playerInitRow, found := data.PlayerInitConfig.Get(req.Gender)
	if !found {
		return nil, fmt.Errorf("player init config not found for gender: %d", req.Gender)
	}

	// 创建角色&添加角色初始的资产
	serverId := session.GetInt32(sessionKey.ServerID)
	newPlayerTable, errCode := db.CreatePlayer(session, req.PlayerName, serverId, playerInitRow)
	if code.IsFail(errCode) {
		return nil, fmt.Errorf("create player failed with code: %d", errCode)
	}

	// TODO 更新最后一次登陆的角色信息到中心节点

	// 抛出角色创建事件
	playerCreateEvent := event.NewPlayerCreate(newPlayerTable.PlayerId, req.PlayerName, req.Gender)
	actor.PostEvent(&playerCreateEvent)

	playerInfo := buildPBPlayer(newPlayerTable)
	response := &msg.PlayerCreateResponse{
		Player: &playerInfo,
	}

	return response, nil
}

// EnterPlayer 进入游戏
func (m *LoginModule) EnterPlayer(session *cproto.Session, req *msg.Int64, actor *pomelo.ActorBase) (*msg.PlayerEnterResponse, error) {
	playerId := req.Value
	if playerId < 1 {
		return nil, fmt.Errorf("invalid player id: %d", playerId)
	}

	// 检查并查找该用户下的该角色
	playerTable, found := db.GetPlayerTable(req.GetValue())
	if !found {
		clog.Warnf("[LoginModule] Player not found in cache: playerId=%d, uid=%d", playerId, session.Uid)
		// 尝试从 UID 查找玩家
		playerIdFromUID := db.GetPlayerIdWithUID(session.Uid)
		if playerIdFromUID > 0 && playerIdFromUID == playerId {
			// 如果通过 UID 能找到对应的 playerId，说明数据存在，可能是缓存问题
			// 尝试重新从缓存获取
			playerTable, found = db.GetPlayerTable(playerId)
			if !found {
				return nil, fmt.Errorf("player not found: %d (even after UID lookup)", playerId)
			}
		} else {
			return nil, fmt.Errorf("player not found: %d", playerId)
		}
	}

	// 保存进入游戏的玩家对应的agentPath
	online.BindPlayer(playerId, playerTable.UID, session.AgentPath)

	// 设置网关节点session的PlayerID属性
	actor.Call(session.ActorPath(), "setSession", &msg.StringKeyValue{
		Key:   sessionKey.PlayerID,
		Value: cstring.ToString(playerId),
	})

	// [99]最后推送 角色进入游戏响应结果
	response := &msg.PlayerEnterResponse{}
	response.GuideMaps = map[int32]int32{}

	// 角色登录事件
	loginEvent := event.NewPlayerLogin(actor.ActorID(), playerId)
	actor.PostEvent(&loginEvent)

	return response, nil
}

// buildPBPlayer 构建玩家信息
func buildPBPlayer(playerTable *db.PlayerTable) msg.Player {
	return msg.Player{
		PlayerId:   playerTable.PlayerId,
		PlayerName: playerTable.Name,
		Level:      playerTable.Level,
		CreateTime: playerTable.CreateTime,
		Exp:        playerTable.Exp,
		Gender:     playerTable.Gender,
	}
}
