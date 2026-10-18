package whatsmeow

import (
	"encoding/json"
	"testing"

	waBinary "github.com/polymorfa/hypermeow/binary"
	"github.com/polymorfa/hypermeow/store"
	"github.com/polymorfa/hypermeow/types"
	"github.com/polymorfa/hypermeow/types/events"
	waLog "github.com/polymorfa/hypermeow/util/log"
)

func TestMessageCappingInfoAcceptsNumbersAndStrings(t *testing.T) {
	var info types.MessageCappingInfo
	raw := `{"capping_status":"FIRST_WARNING","total_quota":1000,"used_quota":"640",` +
		`"cycle_start_timestamp":"1782874800","cycle_end_timestamp":1785553199,"mv_status":"NOT_ACTIVE"}`
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		t.Fatal(err)
	}
	if info.CappingStatus != "FIRST_WARNING" || info.TotalQuota == nil || *info.TotalQuota != 1000 ||
		info.UsedQuota != 640 || info.CycleStart != 1782874800 || info.CycleEnd != 1785553199 || info.MVStatus != "NOT_ACTIVE" {
		t.Fatalf("parsed = %+v", info)
	}
	var none types.MessageCappingInfo
	if err := json.Unmarshal([]byte(`{"capping_status":"NONE"}`), &none); err != nil || none.TotalQuota != nil {
		t.Fatalf("no cap = %+v, %v", none, err)
	}
	if err := json.Unmarshal([]byte(`{"used_quota":"x"}`), &none); err == nil {
		t.Fatal("non-numeric quota must fail")
	}
}

func TestMexNotificationDispatchesMessageCapping(t *testing.T) {
	cli := NewClient(&store.Device{}, waLog.Noop)
	var got *events.MessageCappingUpdate
	cli.AddEventHandler(func(evt any) {
		if e, ok := evt.(*events.MessageCappingUpdate); ok {
			got = e
		}
	})
	payload := `{"data":{"xwa2_notify_new_chat_messages_capping_info_update":{"capping_status":"CAPPED","total_quota":50,"used_quota":50}}}`
	cli.handleMexNotification(t.Context(), &waBinary.Node{
		Tag:   "notification",
		Attrs: waBinary.Attrs{"t": "1782874800"},
		Content: []waBinary.Node{{
			Tag:     "update",
			Attrs:   waBinary.Attrs{"op_name": "MessageCappingInfoNotification"},
			Content: []byte(payload),
		}},
	})
	if got == nil || got.CappingStatus != "CAPPED" || *got.TotalQuota != 50 || got.UsedQuota != 50 ||
		got.Mex.OpName != "MessageCappingInfoNotification" {
		t.Fatalf("event = %+v", got)
	}
}

func TestGroupNodeMemberShareHistoryMode(t *testing.T) {
	cli := NewClient(&store.Device{}, waLog.Noop)
	group, err := cli.parseGroupNode(&waBinary.Node{
		Tag:   "group",
		Attrs: waBinary.Attrs{"id": "123", "creation": "1700000000"},
		Content: []waBinary.Node{
			{Tag: "member_share_group_history_mode", Content: []byte("all_member_share")},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if group.MemberShareHistoryMode != types.GroupMemberShareHistoryModeAllMember {
		t.Fatalf("mode = %q", group.MemberShareHistoryMode)
	}
}

func TestSetGroupMemberShareHistoryModeRejectsUnknownMode(t *testing.T) {
	cli := NewClient(&store.Device{}, waLog.Noop)
	if err := cli.SetGroupMemberShareHistoryMode(t.Context(), types.NewJID("123", types.GroupServer), "everyone"); err == nil {
		t.Fatal("unknown mode must fail before any request")
	}
}
