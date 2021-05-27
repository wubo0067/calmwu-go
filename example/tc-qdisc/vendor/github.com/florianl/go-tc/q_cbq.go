package tc

import (
	"fmt"

	"github.com/mdlayher/netlink"
)

const (
	tcaCbqUnspec = iota
	tcaCbqLssOpt
	tcaCbqWrrOpt
	tcaCbqFOpt
	tcaCbqOVLStrategy
	tcaCbqRate
	tcaCbqRTab
	tcaCbqPolice
)

// Cbq contains attributes of the cbq discipline
type Cbq struct {
	LssOpt      *CbqLssOpt
	WrrOpt      *CbqWrrOpt
	FOpt        *CbqFOpt
	OVLStrategy *CbqOvl
	Rate        *RateSpec
	RTab        []byte
	Police      *CbqPolice
}

// unmarshalCbq parses the Cbq-encoded data and stores the result in the value pointed to by info.
func unmarshalCbq(data []byte, info *Cbq) error {
	ad, err := netlink.NewAttributeDecoder(data)
	if err != nil {
		return err
	}
	ad.ByteOrder = nativeEndian
	for ad.Next() {
		switch ad.Type() {
		case tcaCbqLssOpt:
			arg := &CbqLssOpt{}
			if err := unmarshalStruct(ad.Bytes(), arg); err != nil {
				return err
			}
			info.LssOpt = arg
		case tcaCbqWrrOpt:
			arg := &CbqWrrOpt{}
			if err := unmarshalStruct(ad.Bytes(), arg); err != nil {
				return err
			}
			info.WrrOpt = arg
		case tcaCbqFOpt:
			arg := &CbqFOpt{}
			if err := unmarshalStruct(ad.Bytes(), arg); err != nil {
				return err
			}
			info.FOpt = arg
		case tcaCbqOVLStrategy:
			arg := &CbqOvl{}
			if err := unmarshalStruct(ad.Bytes(), arg); err != nil {
				return err
			}
			info.OVLStrategy = arg
		case tcaCbqRate:
			arg := &RateSpec{}
			if err := unmarshalStruct(ad.Bytes(), arg); err != nil {
				return err
			}
			info.Rate = arg
		case tcaCbqRTab:
			info.RTab = ad.Bytes()
		case tcaCbqPolice:
			arg := &CbqPolice{}
			if err := unmarshalStruct(ad.Bytes(), arg); err != nil {
				return err
			}
			info.Police = arg
		default:
			return fmt.Errorf("unmarshalCbq()\t%d\n\t%v", ad.Type(), ad.Bytes())
		}
	}
	return nil
}

// marshalCbq returns the binary encoding of Qfq
func marshalCbq(info *Cbq) ([]byte, error) {
	options := []tcOption{}

	if info == nil {
		return []byte{}, fmt.Errorf("Cbq: %w", ErrNoArg)
	}
	// TODO: improve logic and check combinations

	if info.LssOpt != nil {
		data, err := marshalStruct(info.LssOpt)
		if err != nil {
			return []byte{}, err
		}
		options = append(options, tcOption{Interpretation: vtBytes, Type: tcaCbqLssOpt, Data: data})
	}
	if info.WrrOpt != nil {
		data, err := marshalStruct(info.WrrOpt)
		if err != nil {
			return []byte{}, err
		}
		options = append(options, tcOption{Interpretation: vtBytes, Type: tcaCbqWrrOpt, Data: data})
	}
	if info.FOpt != nil {
		data, err := marshalStruct(info.FOpt)
		if err != nil {
			return []byte{}, err
		}
		options = append(options, tcOption{Interpretation: vtBytes, Type: tcaCbqFOpt, Data: data})
	}
	if info.OVLStrategy != nil {
		data, err := marshalStruct(info.OVLStrategy)
		if err != nil {
			return []byte{}, err
		}
		options = append(options, tcOption{Interpretation: vtBytes, Type: tcaCbqOVLStrategy, Data: data})
	}
	if info.Police != nil {
		data, err := marshalStruct(info.Police)
		if err != nil {
			return []byte{}, err
		}
		options = append(options, tcOption{Interpretation: vtBytes, Type: tcaCbqPolice, Data: data})
	}

	return marshalAttributes(options)
}

// CbqLssOpt from include/uapi/linux/pkt_sched.h
type CbqLssOpt struct {
	Change  byte
	Flags   byte
	EwmaLog byte
	Level   byte
	Maxidle uint32
	Minidle uint32
	OffTime uint32
	Avpkt   uint32
}

// CbqWrrOpt from include/uapi/linux/pkt_sched.h
type CbqWrrOpt struct {
	Flags     byte
	Priority  byte
	CPriority byte
	Reserved  byte
	Allot     uint32
	Weight    uint32
}

// CbqFOpt from include/uapi/linux/pkt_sched.h
type CbqFOpt struct {
	Split     uint32
	Defmap    uint32
	Defchange uint32
}

// CbqOvl from include/uapi/linux/pkt_sched.h
type CbqOvl struct {
	Strategy  byte
	Priority2 byte
	Pad       uint16
	Penalty   uint32
}

// CbqPolice from include/uapi/linux/pkt_sched.h
type CbqPolice struct {
	Police byte
	Res1   byte
	Res2   uint16
}
