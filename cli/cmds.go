package cli

import (
	cli "github.com/urfave/cli/v2"
)

var Commands = []*cli.Command{
	runCommand,
	tokenSubCommand,
	userSubCommand,
	minerSubCommand,
	signerSubCommand,
}
