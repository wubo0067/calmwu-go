package model

// 数据库表的数据封装
// 对应platform_set下面的device表

const (
	DEVICE_TABLE_NAME = "device"
	DEVICE_DEVICE_ID  = "device_id"
	DEVICE_UIN        = "uin"
	DEVICE_CHANNEL_ID = "channel_id"
	DEVICE_PLATFORM   = "platform"
	DEVICE_RESERVED_0 = "reserved_0"
	DEVICE_RESERVED_1 = "reserved_1"
	DEVICE_RESERVED_2 = "reserved_2"
	DEVICE_RESERVED_3 = "reserved_3"
	DEVICE_RESERVED_4 = "reserved_4"
	DEVICE_RESERVED_5 = "reserved_5"
)

type Device struct {
	DeviceID  string `xorm:"char(128) notnull pk 'device_id'"`
	Uin       int    `xorm:"int index 'uin'"`
	ChannelID string `xorm:"char(128) default('') 'channel_id'"`
	Platform  string `xorm:"char(128) default('') 'platform'"`
	Reserved0 int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1 int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2 string `xorm:"varchar(128) 'reserved_2'"`
	Reserved3 string `xorm:"varchar(1024) 'reserved_3'"`
	Reserved4 string `xorm:"varchar(4096) 'reserved_4'"`
	Reserved5 string `xorm:"text 'reserved_5'"`
}

func (device *Device) TableName() string {
	return DEVICE_TABLE_NAME
}

func (device *Device) GetDeviceFields() []string {
	return []string{DEVICE_DEVICE_ID, DEVICE_UIN, DEVICE_CHANNEL_ID, DEVICE_PLATFORM, DEVICE_RESERVED_0,
		DEVICE_RESERVED_1, DEVICE_RESERVED_2, DEVICE_RESERVED_3, DEVICE_RESERVED_4, DEVICE_RESERVED_5}
}

func (device *Device) QueryDeviceInfo(deviceID string) error {
	return nil
}
