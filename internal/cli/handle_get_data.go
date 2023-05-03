package cli

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v2"
)

func (h *CliHandler) HandleGetData(c *cli.Context) error {
	requestID := c.String("request-id")
	results, err := h.DKGResultByRequestID(requestID)
	if err != nil {
		return err
	}
	filepath := fmt.Sprintf("dkg_results_%s_%d.json", requestID, time.Now().Unix())
	fmt.Printf("writing results to file: %s\n", filepath)
	return writeJSON(filepath, results)
}
