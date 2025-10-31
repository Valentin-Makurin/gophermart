package common

func MapAccrualStatus(accrualStatus string) string {
	switch accrualStatus {
	case "REGISTERED", "PROCESSING":
		return "PROCESSING"
	case "INVALID", "PROCESSED":
		return accrualStatus
	default:
		return "NEW"
	}
}

func IsFinalStatus(status string) bool {
	return status == "PROCESSED" || status == "INVALID"
}
