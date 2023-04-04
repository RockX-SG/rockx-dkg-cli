package storage

import (
	"os"
	"testing"

	"github.com/bloxapp/ssv-spec/types"
	"github.com/stretchr/testify/require"
)

func TestFetchOperatorByID(t *testing.T) {
	os.Setenv("USE_HARDCODED_OPERATORS", "false")

	var testOperatorID types.OperatorID = 1 //LidoRockX

	operator, err := FetchOperatorByID(testOperatorID)

	require.Nil(t, err)
	require.NotNil(t, operator)
}
