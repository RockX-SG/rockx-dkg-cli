/*
 * ==================================================================
 *Copyright (C) 2022-2023 Altstake Technology Pte. Ltd. (RockX)
 *This file is part of rockx-dkg-cli <https://github.com/RockX-SG/rockx-dkg-cli>
 *CAUTION: THESE CODES HAVE NOT BEEN AUDITED
 *
 *rockx-dkg-cli is free software: you can redistribute it and/or modify
 *it under the terms of the GNU General Public License as published by
 *the Free Software Foundation, either version 3 of the License, or
 *(at your option) any later version.
 *
 *rockx-dkg-cli is distributed in the hope that it will be useful,
 *but WITHOUT ANY WARRANTY; without even the implied warranty of
 *MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *GNU General Public License for more details.
 *
 *You should have received a copy of the GNU General Public License
 *along with rockx-dkg-cli. If not, see <http://www.gnu.org/licenses/>.
 *==================================================================
 */

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
