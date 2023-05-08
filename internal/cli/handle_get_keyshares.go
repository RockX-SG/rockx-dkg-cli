package cli

import (
	"fmt"
	"time"

	"github.com/RockX-SG/frost-dkg-demo/internal/utils"
	"github.com/urfave/cli/v2"
)

func (h *CliHandler) HandleGetKeyShares(c *cli.Context) error {
	requestID := c.String("request-id")

	result, err := h.DKGResultByRequestID(requestID)
	if err != nil {
		return fmt.Errorf("HandleGetKeyShares: failed to get dkg result for requestID %s: %w", requestID, err)
	}

	keyshares := &KeyShares{}
	if err := keyshares.ParseDKGResult(result); err != nil {
		return fmt.Errorf("HandleGetKeyShares: failed to parse keyshare from dkg results: %w", err)
	}

	filename := fmt.Sprintf("keyshares-%d.json", time.Now().Unix())
	fmt.Printf("writing keyshares to file: %s\n", filename)
	return utils.WriteJSON(filename, keyshares)
}
