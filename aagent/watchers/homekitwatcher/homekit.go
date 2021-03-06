package homekitwatcher

import (
	"context"
	"crypto/md5"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"

	"github.com/choria-io/go-choria/aagent/util"
	"github.com/choria-io/go-choria/aagent/watchers/event"
)

type State int

const (
	Unknown State = iota
	On
	Off
	// used to indicate that an external event - rpc or other watcher - initiated a transition
	OnNoTransition
	OffNoTransition
)

var stateNames = map[State]string{
	Unknown: "unknown",
	On:      "on",
	Off:     "off",
}

type Machine interface {
	Name() string
	Identity() string
	InstanceID() string
	Version() string
	State() string
	Directory() string
	TimeStampSeconds() int64
	NotifyWatcherState(string, interface{})
	Transition(t string, args ...interface{}) error
	Infof(name string, format string, args ...interface{})
	Errorf(name string, format string, args ...interface{})
}

type transport interface {
	Stop() <-chan struct{}
	Start()
}

type Watcher struct {
	name             string
	states           []string
	failEvent        string
	successEvent     string
	machine          Machine
	statechg         chan struct{}
	previous         State
	interval         time.Duration
	announceInterval time.Duration

	serialNumber  string
	model         string
	pin           string
	setupID       string
	hkt           transport
	ac            *accessory.Switch
	buttonPress   chan State
	initial       State
	path          string
	shouldOn      []string
	shouldOff     []string
	shouldDisable []string
	started       bool

	sync.Mutex
}

func New(machine Machine, name string, states []string, failEvent string, successEvent string, ai time.Duration, properties map[string]interface{}) (watcher *Watcher, err error) {
	w := &Watcher{
		name:             name,
		successEvent:     successEvent,
		failEvent:        failEvent,
		states:           states,
		machine:          machine,
		interval:         5 * time.Second,
		announceInterval: ai,
		model:            "Autonomous Agent",
		statechg:         make(chan struct{}, 1),
		buttonPress:      make(chan State, 1),
		path:             filepath.Join(machine.Directory(), "homekit", fmt.Sprintf("%x", md5.Sum([]byte(name)))),
		shouldOff:        []string{},
		shouldOn:         []string{},
		shouldDisable:    []string{},
	}

	if w.path == "" {
		return nil, fmt.Errorf("machine path could not be determined")
	}

	err = w.setProperties(properties)
	if err != nil {
		return nil, fmt.Errorf("could not set properties: %s", err)
	}

	return w, err
}

func (w *Watcher) Delete() {
	w.Lock()
	defer w.Unlock()

	if w.hkt != nil {
		<-w.hkt.Stop()
	}
}

func (w *Watcher) Type() string {
	return "homekit"
}

func (w *Watcher) AnnounceInterval() time.Duration {
	return w.announceInterval
}

func (w *Watcher) Name() string {
	return w.name
}

func (w *Watcher) NotifyStateChance() {
	if len(w.statechg) < cap(w.statechg) {
		w.statechg <- struct{}{}
	}
}

func (w *Watcher) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	if w.shouldCheck() {
		w.ensureStarted()

		w.machine.Infof(w.name, "homekit watcher for %s starting in state %s", w.name, stateNames[w.initial])
		switch w.initial {
		case On:
			w.buttonPress <- On
		case Off:
			w.buttonPress <- Off
		default:
			w.statechg <- struct{}{}
		}
	}

	for {
		select {
		// button call backs
		case e := <-w.buttonPress:
			err := w.handleStateChange(e)
			if err != nil {
				w.machine.Errorf(w.name, "Could not handle button %s press: %v", stateNames[e], err)
			}

		// rpc initiated state changes wold trigger this
		case <-w.statechg:
			mstate := w.machine.State()
			switch {
			case w.shouldBeOn(mstate):
				w.ensureStarted()
				w.buttonPress <- OnNoTransition

			case w.shouldBeOff(mstate):
				w.ensureStarted()
				w.buttonPress <- OffNoTransition

			case w.shouldBeDisabled(mstate):
				w.ensureStopped()
			}

		case <-ctx.Done():
			w.machine.Infof(w.name, "Stopping on context interrupt")
			w.ensureStopped()
			return
		}
	}
}

func (w *Watcher) ensureStopped() {
	w.Lock()
	defer w.Unlock()

	if !w.started {
		return
	}

	w.machine.Infof(w.name, "Stopping homekit integration")
	<-w.hkt.Stop()
	w.machine.Infof(w.name, "Homekit integration stopped")

	w.started = false
}

func (w *Watcher) ensureStarted() {
	w.Lock()
	defer w.Unlock()

	if w.started {
		return
	}

	w.machine.Infof(w.name, "Starting homekit integration")

	// kind of want to just hk.Start() here but stop kills a context that
	// start does not recreate so we have to go back to start
	err := w.startAccessoryUnlocked()
	if err != nil {
		w.machine.Errorf(w.name, "Could not start homekit service: %s", err)
		return
	}

	go w.hkt.Start()

	w.started = true
}

func (w *Watcher) shouldBeOff(s string) bool {
	for _, state := range w.shouldOff {
		if state == s {
			return true
		}
	}

	return false
}

func (w *Watcher) shouldBeOn(s string) bool {
	for _, state := range w.shouldOn {
		if state == s {
			return true
		}
	}

	return false
}

func (w *Watcher) shouldBeDisabled(s string) bool {
	for _, state := range w.shouldDisable {
		if state == s {
			return true
		}
	}

	return false
}

func (w *Watcher) handleStateChange(s State) error {
	if !w.shouldCheck() {
		return nil
	}

	switch s {
	case On:
		w.setPreviousState(s)
		w.machine.NotifyWatcherState(w.name, w.CurrentState())
		return w.machine.Transition(w.successEvent)

	case OnNoTransition:
		w.setPreviousState(On)
		w.ac.Switch.On.SetValue(true)
		w.machine.NotifyWatcherState(w.name, w.CurrentState())
		return nil

	case Off:
		w.setPreviousState(s)
		w.machine.NotifyWatcherState(w.name, w.CurrentState())
		return w.machine.Transition(w.failEvent)

	case OffNoTransition:
		w.setPreviousState(Off)
		w.ac.Switch.On.SetValue(false)
		w.machine.NotifyWatcherState(w.name, w.CurrentState())
		return nil
	}

	return fmt.Errorf("invalid state change event: %s", stateNames[s])
}

func (w *Watcher) CurrentState() interface{} {
	w.Lock()
	defer w.Unlock()

	s := &StateNotification{
		Event: event.Event{
			Protocol:  "io.choria.machine.watcher.homekit.v1.state",
			Type:      "homekit",
			Name:      w.name,
			Identity:  w.machine.Identity(),
			ID:        w.machine.InstanceID(),
			Version:   w.machine.Version(),
			Timestamp: w.machine.TimeStampSeconds(),
			Machine:   w.machine.Name(),
		},
		Path:            w.path,
		PreviousOutcome: stateNames[w.previous],
	}

	return s
}

func (w *Watcher) setPreviousState(s State) {
	w.Lock()
	defer w.Unlock()

	w.previous = s
}

func (w *Watcher) startAccessoryUnlocked() error {
	info := accessory.Info{
		Name:             strings.Title(strings.Replace(w.name, "_", " ", -1)),
		SerialNumber:     w.serialNumber,
		Manufacturer:     "Choria",
		Model:            w.model,
		FirmwareRevision: w.machine.Version(),
	}
	w.ac = accessory.NewSwitch(info)

	t, err := hc.NewIPTransport(hc.Config{Pin: w.pin, SetupId: w.setupID, StoragePath: w.path}, w.ac.Accessory)
	if err != nil {
		return err
	}

	hc.OnTermination(func() {
		<-t.Stop()
	})

	w.ac.Switch.On.OnValueRemoteUpdate(func(new bool) {
		w.Lock()
		defer w.Unlock()

		w.machine.Infof(w.name, "Handling app button press: %v", new)

		if !w.shouldCheck() {
			w.machine.Infof(w.name, "Ignoring event while in %s state", w.machine.State())
			// undo the button press
			w.ac.Switch.On.SetValue(!new)
			return
		}

		if new {
			w.machine.Infof(w.name, "Setting state to On")
			w.buttonPress <- On
		} else {
			w.machine.Infof(w.name, "Setting state to Off")
			w.buttonPress <- Off
		}
	})

	w.ac.Switch.On.SetValue(w.previous == On)

	w.hkt = t

	return nil
}

func (w *Watcher) shouldCheck() bool {
	if len(w.states) == 0 {
		return true
	}

	for _, e := range w.states {
		if e == w.machine.State() {
			return true
		}
	}

	return false
}

func (w *Watcher) validate() error {
	if len(w.pin) > 0 && len(w.pin) != 8 {
		return fmt.Errorf("pin should be 8 characters long")
	}

	if len(w.setupID) > 0 && len(w.setupID) != 4 {
		return fmt.Errorf("setup_id should be 4 characters long")
	}

	return nil
}

func (w *Watcher) setProperties(props map[string]interface{}) error {
	var properties struct {
		SerialNumber  string `mapstructure:"serial_number"`
		Model         string
		Pin           string
		SetupId       string   `mapstructure:"setup_id"`
		ShouldOn      []string `mapstructure:"on_when"`
		ShouldOff     []string `mapstructure:"off_when"`
		ShouldDisable []string `mapstructure:"disable_when"`
		Initial       bool
	}

	err := util.ParseMapStructure(props, &properties)
	if err != nil {
		return err
	}

	w.serialNumber = properties.SerialNumber
	w.model = properties.Model
	w.pin = properties.Pin
	w.setupID = properties.SetupId
	w.shouldOn = properties.ShouldOn
	w.shouldOff = properties.ShouldOff
	w.shouldDisable = properties.ShouldDisable

	if properties.Initial {
		w.initial = On
	} else {
		w.initial = Off
	}

	return w.validate()
}
