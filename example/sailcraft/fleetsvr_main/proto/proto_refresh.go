package proto

type ProtoRefreshResponse struct {
	ActivityTaskFreshed         int `json:"ActivityTaskFreshed"`
	ActivityTaskRestTimeToFresh int `json:"ActivityTaskRestTimeToFresh"`
}
