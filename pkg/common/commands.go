package common

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

var commands []*cli.Command

// RegisterCommand --
func RegisterCommand(command *cli.Command) {
	logrus.Debugln("Registering", command.Name, "command...")
	commands = append(commands, command)
}

// GetCommands --
func GetCommands() []*cli.Command {
	return commands
}

func CheckFilePath(ctx context.Context, _ *cli.Command, s string) error {
	timeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	check := make(chan error)
	go func() {
		_, err := os.Stat(s)
		check <- err
	}()

	select {
	case <-timeout.Done():
		return timeout.Err()
	case err := <-check:
		return err
	}
}

func CheckRealInt(_ context.Context, _ *cli.Command, i int) error {
	if i > math.MaxInt || i < 0 {
		return fmt.Errorf("value must be between 0 and %d", math.MaxInt)
	}
	return nil
}
