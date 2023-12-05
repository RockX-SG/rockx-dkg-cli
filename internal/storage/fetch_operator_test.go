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

package storage

import (
	"os"
	"testing"

	"github.com/bloxapp/ssv-spec/types"
	"github.com/stretchr/testify/require"
)

func TestFetchOperatorByID(t *testing.T) {
	os.Setenv("USE_HARDCODED_OPERATORS", "false")

	networks := []string{"mainnet", "jato-v2", "prater", "goerli", "holesky"}
	testOperatorIDs := []types.OperatorID{18, 31, 31, 31, 119}

	for i, network := range networks {
		t.Run(network, func(t *testing.T) {
			os.Setenv("OPERATOR_REGISTRY_NETWORK", network)
			operator, err := FetchOperatorByID(testOperatorIDs[i])
			require.Nil(t, err)
			require.NotNil(t, operator)
			require.Equal(t, testOperatorIDs[i], operator.OperatorID)
		})
	}
}

func TestOperatorRegistryNetwork(t *testing.T) {
	networks := []struct {
		Network  string
		Expected string
	}{
		{
			Network:  "mainnet",
			Expected: "mainnet",
		},
		{
			Network:  "jato-v2",
			Expected: "prater",
		},
		{
			Network:  "prater",
			Expected: "prater",
		},
		{
			Network:  "goerli",
			Expected: "prater",
		},
		{
			Network:  "holesky",
			Expected: "holesky",
		},
	}

	for _, network := range networks {
		t.Run(network.Network, func(t *testing.T) {
			os.Setenv("OPERATOR_REGISTRY_NETWORK", network.Network)
			got := OperatorRegistryNetwork()

			require.Equal(t, network.Expected, got)
		})
	}

}
