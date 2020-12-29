package proto

const GUILD_ID_INVALID = "invalid"

func NewDefaultProtoUserInfoChanged() *ProtoUserInfo {
	userInfochanged := new(ProtoUserInfo)
	userInfochanged.Level = -1
	userInfochanged.Exp = -1
	userInfochanged.Star = -1
	userInfochanged.Gold = -1
	userInfochanged.Wood = -1
	userInfochanged.Gem = -1
	userInfochanged.PurchaseGem = -1
	userInfochanged.Stone = -1
	userInfochanged.Iron = -1
	userInfochanged.ChangeNameCount = -1
	userInfochanged.GuildId = GUILD_ID_INVALID

	return userInfochanged
}
