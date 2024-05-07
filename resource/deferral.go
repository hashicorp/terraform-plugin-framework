package resource

const (
	DeferralReasonUnknown               DeferredReason = 0
	DeferralReasonResourceConfigUnknown DeferredReason = 1
	DeferralReasonProviderConfigUnknown DeferredReason = 2
	DeferralReasonAbsentPrereq          DeferredReason = 3
)

type DeferredResponse struct {
	Reason DeferredReason
}

type DeferredReason int32

func (d DeferredReason) String() string {
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
