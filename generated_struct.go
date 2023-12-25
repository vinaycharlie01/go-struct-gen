package main

type EventData struct {
	Name string `json:"name,omitempty"`
	Status string `json:"status,omitempty"`
	Key string `json:"Key,omitempty"`
	Event Event `json:"Event,omitempty"`
}
type Comp struct {
			CompType string `json:"CompType,omitempty"`
			CompTypeInt int `json:"CompTypeInt,omitempty"`
			CompId int `json:"CompId,omitempty"`
		}
type PrimaryComp struct {
		CompType string `json:"CompType,omitempty"`
		CompTypeInt int `json:"CompTypeInt,omitempty"`
		CompId int `json:"CompId,omitempty"`
		Comp Comp `json:"Comp,omitempty"`
		}
type Tier struct {
				TierType string `json:"TierType,omitempty"`
				TierInt int `json:"TierInt,omitempty"`
			}
type MaintModeReason struct {
					MMReason string `json:"MMReason,omitempty"`
					MMInt int `json:"MMInt,omitempty"`
					MMLongRun string `json:"MMLongRun,omitempty"`
				}
type Event struct {
	Node int `json:"Node,omitempty"`
	Seq int `json:"Seq,omitempty"`
	Eincn int `json:"Eincn,omitempty"`
	Sev int `json:"Sev,omitempty"`
	Type string `json:"Type,omitempty"`
	Time int `json:"Time,omitempty"`
	HiResTime int64 `json:"HiResTime,omitempty"`
	MessageCode int `json:"MessageCode,omitempty"`
	AlertId int `json:"AlertId,omitempty"`
	Class string `json:"Class,omitempty"`
	PrimaryComp PrimaryComp `json:"PrimaryComp,omitempty"`
			ComponentKey string `json:"ComponentKey,omitempty"`
			Tier Tier `json:"Tier,omitempty"`
				AlertTypeInt int `json:"AlertTypeInt,omitempty"`
				MaintModeReason MaintModeReason `json:"MaintModeReason,omitempty"`
					MsgParams string `json:"MsgParams,omitempty"`
					NonscAlertMetaInt int `json:"NonscAlertMetaInt,omitempty"`
					ShortDesc string `json:"ShortDesc,omitempty"`
					EventString string `json:"EventString,omitempty"`
					Serial string `json:"Serial,omitempty"`
					SystemModel string `json:"SystemModel,omitempty"`
					ReleaseLevel string `json:"ReleaseLevel,omitempty"`
				}