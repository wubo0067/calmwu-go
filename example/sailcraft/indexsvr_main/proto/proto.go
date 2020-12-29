/*
 * @Author: calmwu
 * @Date: 2017-09-19 09:59:54
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-28 15:27:43
 * @Comment:
 */

package proto

type DataSetType int

const (
	CTRLCMD_REOLADCONF_REQ = "reload_conf_req"
	CTRLCMD_REOLADCONF_RES = "reload_conf_res"
	CTRLCMD_RELOADDATA_REQ = "reload_data_req"
	CTRLCMD_RELOADDATA_RES = "reload_data_res"
)

const (
	E_DATATYPE_GUILDINFO = iota
	E_DATATYPE_USERINFO
)

type DataMetaI interface {
	Type() DataSetType
	Compare(object DataMetaI) bool
}

type UserInfoS struct {
	Uin      string `mapstructure:"uin"`
	UserName string `mapstructure:"user_name"`
}

type GuildInfoS struct {
	ID        string `mapstructure:"id"`
	GuildName string `mapstructure:"name"`
	Creator   string `mapstructure:"creator"`
	PerformId string `mapstructure:"perform_id"`
}

func (user *UserInfoS) Type() DataSetType {
	return E_DATATYPE_USERINFO
}

func (user *UserInfoS) Compare(object DataMetaI) bool {
	if other, ok := object.(*UserInfoS); ok {
		if other.Uin == user.Uin && other.Type() == E_DATATYPE_USERINFO {
			return true
		}
	}
	return false
}

func (guild *GuildInfoS) Type() DataSetType {
	return E_DATATYPE_GUILDINFO
}

func (guild *GuildInfoS) Compare(object DataMetaI) bool {
	if other, ok := object.(*GuildInfoS); ok {
		if other.ID == guild.ID &&
			other.Type() == E_DATATYPE_GUILDINFO &&
			other.Creator == guild.Creator {
			return true
		}
	}
	return false
}

type ProtoFindGuildsByNameRequestParamsS struct {
	GuildName  string `json:"GuildName"`
	QueryType  string `json:"QueryType"`
	QueryCount int    `json:"QueryCount"`
}

type ProtoFindGuildsByNameResponseParamsS struct {
	GuildCount int           `json:"GuildCount"`
	GuildInfos []*GuildInfoS `json:"GuildInfos"`
}

type ProtoFindUsersByNameRequestParamsS struct {
	UserName   string `json:"UserName"`
	QueryType  string `json:"QueryType"`
	QueryCount int    `json:"QueryCount"`
}

type ProtoFindUsersByNameResponseParamsS struct {
	UserCount int          `json:"UserCount"`
	UserInfos []*UserInfoS `json:"UserInfos"`
}

type ProtoModifyUserNameReqParamsS struct {
	UserName    string `json:"UserName"`
	Uin         string `json:"Uin"`
	NewUserName string `json:"NewUserName"`
}

type ProtoModifyGuildNameReqParamsS struct {
	GuildName    string `json:"GuildName"`
	ID           string `json:"ID"`
	NewGuildName string `json:"NewUserName"`
}

type ProtoDeleteGuildNameReqParamsS struct {
	GuildName string `json:"GuildName"`
	ID        string `json:"ID"`
	Creator   string `json:"Creator"`
	PerformId string `json:"PerformId"`
}

type ProtoAddGuildNameReqParamsS ProtoDeleteGuildNameReqParamsS

type ProtoAddUserNameReqParamsS struct {
	Uin      string `json:"uin"`
	UserName string `json:"user_name"`
}

type ProtoControlCmdS struct {
	CmdName string `json:"CmdName"`
	CmdData string `json:"CmdData"`
}
type ProtoDirtyWordFilterReq struct {
	Uin     uint64 `json:"Uin"`
	Content string `json:"Content"`
}

type ProtoDirtyWordFilterRes struct {
	Uin            uint64 `json:"Uin"`
	HaveDirtyWords int    `json:"HaveDirtyWords"` // 1: 有dirty内容 0：没有dirty内容
	FilterContent  string `json:"FilterContent"`  // 过滤后内容
}
