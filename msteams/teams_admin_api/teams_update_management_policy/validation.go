package teamsupdatemanagementpolicy

func validateSettings(v ScalarSettings) error {
	if n, ok := v.UpdateDayOfWeek.Value(); ok && (n < 0 || n > 6) {
		return ErrInput
	}
	if date, ok := v.UpdateTimeOfDay.Value(); ok {
		if _, e := date.MarshalJSON(); e != nil {
			return ErrInput
		}
	}
	return nil
}
