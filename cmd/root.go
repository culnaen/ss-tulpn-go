package cmd

import (
	"fmt"

	"github.com/culnaen/ss-tulpn-go/internal/proc"
)

func Execute() error {
	user_entities, err := proc.GetUserEntities()
	if err != nil {
		return err
	}
	fmt.Printf("%-*s ", 10, "State")
	fmt.Printf("%-6s %-6s ", "Recv-Q", "Send-Q")
	fmt.Printf("%*s:%-*s %*s:%-*s %-*s\n",
		15, "Local Address", 5, "Port",
		15, "Peer Address", 5, "Port",
		10, "Process",
	)

	if err := proc.ShowNetUdp(user_entities); err != nil {
		return err
	}

	if err := proc.ShowNetTcp(user_entities); err != nil {
		return err
	}

	return nil
}
