package cmd

import (
	"sync"

	"github.com/sirupsen/logrus"

	scoutcmd "github.com/choria-io/go-choria/scout/cmd"
)

type sTriggerCommand struct {
	identities []string
	facts      []string
	classes    []string
	checks     []string
	combined   []string
	json       bool
	verbose    bool

	command
}

func (s *sTriggerCommand) Setup() (err error) {
	if scout, ok := cmdWithFullCommand("scout"); ok {
		s.cmd = scout.Cmd().Command("trigger", "Trigger immediate check invocations")
		s.cmd.Flag("wf", "Match hosts with a certain fact").Short('F').StringsVar(&s.facts)
		s.cmd.Flag("wc", "Match hosts with a certain configuration management class").Short('C').StringsVar(&s.classes)
		s.cmd.Flag("wi", "Match hosts with a certain Choria identity").Short('I').StringsVar(&s.identities)
		s.cmd.Flag("with", "Combined classes and facts filter").Short('W').PlaceHolder("FILTER").StringsVar(&s.combined)
		s.cmd.Flag("check", "Trigger only specific checks").StringsVar(&s.checks)
		s.cmd.Flag("json", "JSON format output").BoolVar(&s.json)
		s.cmd.Flag("verbose", "Show verbose output").Short('v').BoolVar(&s.verbose)
	}

	return nil
}

func (s *sTriggerCommand) Configure() error {
	return commonConfigure()
}

func (s *sTriggerCommand) Run(wg *sync.WaitGroup) (err error) {
	defer wg.Done()

	trigger, err := scoutcmd.NewTriggerCommand(s.identities, s.classes, s.facts, s.combined, s.checks, s.json, configFile, debug || s.verbose, logrus.NewEntry(c.Logger("scout").Logger))
	if err != nil {
		return err
	}

	wg.Add(1)
	return trigger.Run(ctx, wg)
}

func init() {
	cli.commands = append(cli.commands, &sTriggerCommand{})
}
