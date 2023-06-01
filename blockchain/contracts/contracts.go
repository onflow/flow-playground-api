package contracts

import (
	"embed"
	"fmt"
)

// Embed all contracts in this folder
//
//go:embed *.cdc
var contracts embed.FS

func Get(name string) ([]byte, error) {
	return contracts.ReadFile(fmt.Sprintf("%s.cdc", name))
}
