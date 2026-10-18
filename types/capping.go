package types

import (
	"fmt"
	"strconv"
	"strings"
)

// FlexInt is an integer that WhatsApp may send as a JSON number or as a string.
type FlexInt int64

func (f *FlexInt) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		*f = 0
		return nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid integer %s: %w", b, err)
	}
	*f = FlexInt(n)
	return nil
}

// MessageCappingInfo is the account's quota of messages that start new chats, per cycle
// (xwa2_message_capping_info). The status values are an open set: WhatsApp may add new ones.
type MessageCappingInfo struct {
	CappingStatus string   `json:"capping_status"` // NONE, FIRST_WARNING, SECOND_WARNING, CAPPED
	TotalQuota    *FlexInt `json:"total_quota"`    // nil when the account has no cap
	UsedQuota     FlexInt  `json:"used_quota"`
	CycleStart    FlexInt  `json:"cycle_start_timestamp"` // unix seconds; 0 when unknown
	CycleEnd      FlexInt  `json:"cycle_end_timestamp"`
	MVStatus      string   `json:"mv_status"`
	OTEStatus     string   `json:"ote_status"`
}
