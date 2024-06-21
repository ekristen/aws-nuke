package nuke

import (
	"fmt"
	"time"

	libnuke "github.com/ekristen/libnuke/pkg/nuke"
	"github.com/ekristen/libnuke/pkg/utils"

	"github.com/ekristen/aws-nuke/v3/pkg/awsutil"
)

// Prompt struct provides a way to provide a custom prompt to the libnuke library this allows
// custom data to be available to the Prompt func when it's executed by the libnuke library
type Prompt struct {
	Parameters *libnuke.Parameters
	Account    *awsutil.Account
}

// Prompt is the actual function called by the libnuke process during it's run
func (p *Prompt) Prompt() error {
	forceSleep := time.Duration(p.Parameters.ForceSleep) * time.Second

	fmt.Printf("Do you really want to nuke the account with "+
		"the ID %s and the alias '%s'?\n", p.Account.ID(), p.Account.Alias())
	if p.Parameters.Force {
		fmt.Printf("Waiting %v before continuing.\n", forceSleep)
		time.Sleep(forceSleep)
	} else {
		fmt.Printf("Do you want to continue? Enter account alias to continue.\n")
		if err := utils.Prompt(p.Account.Alias()); err != nil {
			return err
		}
	}

	return nil
}
