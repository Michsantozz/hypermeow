package whatsmeow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/polymorfa/hypermeow/types"
)

const mutationUpdateGroupProperty = "9418211574894172"

// SetGroupMemberShareHistoryMode sets who may send the group's recent messages to new members:
// 'admin_share' (admins only) or 'all_member_share' (any member).
func (cli *Client) SetGroupMemberShareHistoryMode(ctx context.Context, jid types.JID, mode types.GroupMemberShareHistoryMode) error {
	var value string
	switch mode {
	case types.GroupMemberShareHistoryModeAdmin:
		value = "ADMIN_SHARE"
	case types.GroupMemberShareHistoryModeAllMember:
		value = "ALL_MEMBER_SHARE"
	default:
		return errors.New("invalid mode, must be 'admin_share' or 'all_member_share'")
	}
	data, err := cli.sendMexIQ(ctx, mutationUpdateGroupProperty, map[string]any{
		"group_id": jid.String(),
		"update":   map[string]any{"member_share_group_history_mode": value},
	})
	if err != nil {
		return err
	}
	var resp struct {
		Result *struct {
			State string `json:"state"`
		} `json:"xwa2_group_update_property"`
	}
	if err = json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("failed to parse group property response: %w", err)
	}
	// WhatsApp Web only fails when a state is present and is not ACTIVE.
	if resp.Result != nil && resp.Result.State != "" && resp.Result.State != "ACTIVE" {
		return fmt.Errorf("failed to update group property: state %s", resp.Result.State)
	}
	return nil
}
