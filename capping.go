package whatsmeow

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/polymorfa/hypermeow/types"
)

const queryMessageCappingInfo = "24503548349331633"

// GetMessageCappingInfo fetches the account's quota of messages that start new chats. WhatsApp
// pushes later changes as events.MessageCappingUpdate, but not on connect: fetch it once then.
func (cli *Client) GetMessageCappingInfo(ctx context.Context) (*types.MessageCappingInfo, error) {
	data, err := cli.sendMexIQ(ctx, queryMessageCappingInfo, map[string]any{
		"input": map[string]any{"type": "INDIVIDUAL_NEW_CHAT_MSG"},
	})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Info *types.MessageCappingInfo `json:"xwa2_message_capping_info"`
	}
	if err = json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse message capping info: %w", err)
	} else if resp.Info == nil {
		return nil, &ElementMissingError{Tag: "xwa2_message_capping_info", In: "mex response"}
	}
	return resp.Info, nil
}
