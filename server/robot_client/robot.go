package main

import (
	"fmt"
	"math/rand"
	"time"

	"lucky/server/gen/msg"
	"lucky/server/pkg/code"

	cherryError "github.com/cherry-game/cherry/error"
	cherryHttp "github.com/cherry-game/cherry/extend/http"
	cherryTime "github.com/cherry-game/cherry/extend/time"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryClient "github.com/cherry-game/cherry/net/parser/pomelo/client"
	jsoniter "github.com/json-iterator/go"
)

type (
	// Robot client robot
	Robot struct {
		*cherryClient.Client
		PrintLog   bool
		Token      string
		ServerId   int32
		PID        int32
		UID        int64
		OpenId     string
		PlayerId   int64
		PlayerName string
		StartTime  cherryTime.CherryTime
	}
)

func New(client *cherryClient.Client) *Robot {
	return &Robot{
		Client: client,
	}
}

// GetToken  http登录获取token对象
// http://172.16.124.137/login?pid=2126003&account=test1&password=test1
func (p *Robot) GetToken(url string, pid, userName, password string) error {
	// http登陆获取token json对象
	requestURL := fmt.Sprintf("%s/login", url)
	jsonBytes, _, err := cherryHttp.GET(requestURL, map[string]string{
		"pid":      pid,      //sdk包id
		"account":  userName, //帐号名
		"password": password, //密码
	})

	if err != nil {
		return err
	}

	// 转换json对象
	rsp := code.Result{}
	if err = jsoniter.Unmarshal(jsonBytes, &rsp); err != nil {
		return err
	}

	if code.IsFail(rsp.Code) {
		return cherryError.Errorf("get Token fail. [message = %s]", rsp.Message)
	}

	// 获取token值
	p.Token = rsp.Data.(string)
	p.TagName = fmt.Sprintf("%s_%s", pid, userName)
	p.StartTime = cherryTime.Now()

	return nil
}

// UserLogin 用户登录对某游戏服
func (p *Robot) UserLogin(serverId int32) error {
	route := "gate.user.login"

	p.Debugf("[%s] [UserLogin] request ServerID = %d", p.TagName, serverId)

	msg, err := p.Request(route, &msg.LoginRequest{
		ServerId: serverId,
		Token:    p.Token,
		Params:   nil,
	})

	if err != nil {
		return err
	}

	p.ServerId = serverId

	rsp := &msg.LoginResponse{}
	err = p.Serializer().Unmarshal(msg.Data, rsp)
	if err != nil {
		return err
	}

	p.UID = rsp.Uid
	p.PID = rsp.Pid
	p.OpenId = rsp.OpenId

	p.Debugf("[%s] [UserLogin] response = %+v", p.TagName, rsp)
	return nil
}

// PlayerSelect 查看玩家列表
func (p *Robot) PlayerSelect() error {
	route := "game.player.select"

	msg, err := p.Request(route, &msg.None{})
	if err != nil {
		return err
	}

	rsp := &msg.PlayerSelectResponse{}
	err = p.Serializer().Unmarshal(msg.Data, rsp)
	if err != nil {
		return err
	}

	if len(rsp.List) < 1 {
		p.Debugf("[%s] not found player list.", p.TagName)
		return nil
	}

	p.PlayerId = rsp.List[0].PlayerId
	p.PlayerName = rsp.List[0].PlayerName

	p.Debugf("[%s] [PlayerSelect] response PlayerID = %d,PlayerName = %s", p.TagName, p.PlayerId, p.PlayerName)

	return nil
}

// ActorCreate 创建角色（使用默认名称）
func (p *Robot) ActorCreate() error {
	return p.ActorCreateWithName("p" + p.OpenId)
}

// ActorCreateWithName 创建角色（指定名称）
func (p *Robot) ActorCreateWithName(playerName string) error {
	if p.PlayerId > 0 {
		p.Debugf("[%s] deny create actor", p.TagName)
		return nil
	}

	route := "game.player.create"
	gender := rand.Int31n(1)

	req := &msg.PlayerCreateRequest{
		PlayerName: playerName,
		Gender:     gender,
	}

	msg, err := p.Request(route, req)
	if err != nil {
		return err
	}

	rsp := &msg.PlayerCreateResponse{}
	err = p.Serializer().Unmarshal(msg.Data, rsp)
	if err != nil {
		return err
	}

	p.PlayerId = rsp.Player.PlayerId
	p.PlayerName = rsp.Player.PlayerName

	p.Debugf("[%s] [ActorCreate] PlayerID = %d,ActorName = %s", p.TagName, p.PlayerId, p.PlayerName)

	return nil
}

// ActorEnter 角色进入游戏
func (p *Robot) ActorEnter() error {
	route := "game.player.enter"
	req := &msg.Int64{
		Value: p.PlayerId,
	}

	msg, err := p.Request(route, req)
	if err != nil {
		return err
	}

	rsp := &msg.PlayerEnterResponse{}
	err = p.Serializer().Unmarshal(msg.Data, rsp)
	if err != nil {
		return err
	}

	p.Debugf("[%s] [ActorEnter] response PlayerID = %d,ActorName = %s", p.TagName, p.PlayerId, p.PlayerName)
	return nil
}

// BuyItem 购买道具
func (p *Robot) BuyItem(shopId, itemId, count, payType int32) error {
	route := "game.player.buyItem"

	req := &msg.BuyItemRequest{
		ShopId:  shopId,
		ItemId:  itemId,
		Count:   count,
		PayType: payType,
	}

	p.Debugf("[%s] [BuyItem] request shopId=%d, itemId=%d, count=%d, payType=%d",
		p.TagName, shopId, itemId, count, payType)

	msg, err := p.Request(route, req)
	if err != nil {
		return err
	}

	rsp := &msg.BuyItemResponse{}
	err = p.Serializer().Unmarshal(msg.Data, rsp)
	if err != nil {
		return err
	}

	p.Debugf("[%s] [BuyItem] response itemId=%d, count=%d, payType=%d, costAmount=%d, items=%+v",
		p.TagName, rsp.ItemId, rsp.Count, rsp.PayType, rsp.CostAmount, rsp.Items)

	return nil
}

func (p *Robot) RandSleep() {
	time.Sleep(time.Duration(rand.Int31n(10)) * time.Millisecond)
}

func (p *Robot) Debug(args ...interface{}) {
	if p.PrintLog {
		cherryLogger.Debug(args...)
	}

}

func (p *Robot) Debugf(template string, args ...interface{}) {
	if p.PrintLog {
		cherryLogger.Debugf(template, args...)
	}
}
