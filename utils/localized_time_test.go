package utils

import (
	"testing"
)

// 1522222153 2018-03-28 周三
// 1522308553 2018-03-29 周四
// 1522826953 2018-04-04 下周三
func TestIsSameWeek(t *testing.T) {
	if !GLocalizedTime.IsTheSameWeek(1522222153, 1522308553) {
		t.Errorf("[IsSameWeek] got false expected true")
	}

	if GLocalizedTime.IsTheSameWeek(1522222153, 1522826953) {
		t.Errorf("[IsSameWeek] got true expected false")
	}
}
