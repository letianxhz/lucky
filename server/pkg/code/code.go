package code

var (
	OK                       int32 = 0   // is ok
	Error                    int32 = 1   // error
	PIDError                 int32 = 100 // pid错误
	LoginError               int32 = 201 // 登录异常
	AccountAuthFail          int32 = 202 // 帐号授权失败
	AccountBindFail          int32 = 203 // 帐号绑定失败
	AccountTokenValidateFail int32 = 204 // token验证失败
	AccountNameIsExist       int32 = 205 // 帐号已存在
	AccountRegisterError     int32 = 206 //
	AccountGetFail           int32 = 207 //
	PlayerDenyLogin          int32 = 301 // 玩家禁止登录
	PlayerDuplicateLogin     int32 = 302 // 玩家重复登录
	PlayerNameExist          int32 = 303 // 玩家角色名已存在
	PlayerCreateFail         int32 = 304 // 玩家创建角色失败
	PlayerNotLogin           int32 = 305 // 玩家未登录
	PlayerIDError            int32 = 306 // 玩家id错误
	ShopItemNotFound         int32 = 401 // 商店道具不存在
	ShopItemNotEnoughMoney   int32 = 402 // 货币不足
	ShopItemInvalidParam     int32 = 403 // 购买参数错误
	ShopItemBuyFail          int32 = 404 // 购买失败
)
