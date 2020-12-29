package handler

import (
	"sailcraft/base"
	"sailcraft/fleetsvr_main/handlerbase"
)

type TimerHandlerFunc func(*base.ProtoRequestS, *map[string]interface{}) (int, error)

var timerHandlers []TimerHandlerFunc

func init() {
	RegisterTimerHandler(OnCampaignEventTriggerForTimer)
}

type TimerHandler struct {
	handlerbase.WebHandler
}

func (this *TimerHandler) OnTimerTiming() (int, error) {
	resParams := make(map[string]interface{})

	for _, timerHandler := range timerHandlers {
		if timerHandler != nil {
			retCode, err := timerHandler(this.Request, &resParams)
			if err != nil {
				return retCode, err
			}
		}
	}

	this.Response.ResData.Params = resParams
	return 0, nil
}

func RegisterTimerHandler(hd TimerHandlerFunc) {
	if hd != nil {
		timerHandlers = append(timerHandlers, hd)
	}
}
