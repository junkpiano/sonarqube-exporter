package internal

func ConvertStatusToFloat(s string) float64 {
	if s == "GREEN" {
		return 1.0
	} else {
		return 0.0
	}
}
