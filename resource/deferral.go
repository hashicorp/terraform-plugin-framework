package resource

const (
	DeferralReasonUnknown               DeferralReason = 0
	DeferralReasonResourceConfigUnknown DeferralReason = 1
	DeferralReasonProviderConfigUnknown DeferralReason = 2
	DeferralReasonAbsentPrereq          DeferralReason = 3
)

type DeferralResponse struct {
	Reason DeferralReason
}

type DeferralReason int32

func (d DeferralReason) String() string {
	switch d {
	case 0:
		return "Unknown"
	case 1:
		return "Resource Config Unknown"
	case 2:
		return "Provider Config Unknown"
	case 3:
		return "Absent Prerequisite"
	}
	return "Unknown"
}
