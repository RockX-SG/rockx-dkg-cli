package cli

import (
	"fmt"
	"time"

	"github.com/RockX-SG/frost-dkg-demo/internal/utils"
	"github.com/urfave/cli/v2"
)

func (h *CliHandler) HandleGetData(c *cli.Context) error {
	requestID := c.String("request-id")
	results, err := h.DKGResultByRequestID(requestID)
	if err != nil {
		return fmt.Errorf("HandleGetData: failed to get dkg result for requestID %s: %w", requestID, err)
	}
	filepath := fmt.Sprintf("dkg_results_%s_%d.json", requestID, time.Now().Unix())
	fmt.Printf("writing results to file: %s\n", filepath)
	return utils.WriteJSON(filepath, results)
}
