package faaschain

import (
	"github.com/s8sg/faaschain/sdk"
	"net/url"
	"path"
)

type Options struct {
	header map[string]string
	query  map[string][]string
	sync   bool
}

type Fchain struct {
	chain    *sdk.Chain
	id       string
	url      string
	asyncUrl string
	chainDef []byte
}

type Option func(*Options)

var (
	// Sync can be used instead of SyncCall
	Sync = SyncCall()
	// Denote if last phase doesn't contain any function
	emptyPhase = false
)

func (o *Options) reset() {
	o.header = map[string]string{}
	o.query = map[string][]string{}
	o.sync = false
}

// SyncCall Set sync flag
func SyncCall() Option {
	return func(o *Options) {
		o.sync = true
	}
}

// Header Specify a header
func Header(key, value string) Option {
	return func(o *Options) {
		o.header[key] = value
	}
}

// Query Specify a query parameter
func Query(key string, value ...string) Option {
	return func(o *Options) {
		array := []string{}
		for _, val := range value {
			array = append(array, val)
		}
		o.query[key] = array
	}
}

// NewFaaschain creates a new faaschain object
func NewFaaschain(gateway string, chain string) *Fchain {
	fchain := &Fchain{}
	fchain.chain = sdk.CreateChain()
	u, _ := url.Parse(gateway)
	u.Path = path.Join(u.Path, "function/"+chain)
	fchain.url = u.String()
	u, _ = url.Parse(gateway)
	u.Path = path.Join(u.Path, "async-function/"+chain)
	fchain.asyncUrl = u.String()
	return fchain
}

// SetId set request id to a chain
func (fchain *Fchain) SetId(id string) {
	fchain.id = id
}

// ApplyModifier allows to apply inline callback functionl
// the callback fucntion prototype is
// func([]byte) ([]byte, error)
func (fchain *Fchain) ApplyModifier(mod sdk.Modifier) *Fchain {
	var phase *sdk.Phase
	newMod := sdk.CreateModifier(mod)
	if len(fchain.chain.Phases) == 0 {
		phase = sdk.CreateExecutionPhase()
		fchain.chain.AddPhase(phase)
		emptyPhase = true
	} else {
		phase = fchain.chain.GetLastPhase()
	}
	phase.AddFunction(newMod)
	return fchain
}

// Callback register a callback url as a part of chain definition
// One or more callback function can be placed for sending
// either partial chain data or after the chain completion
func (fchain *Fchain) Callback(url string, opts ...Option) *Fchain {
	newCallback := sdk.CreateCallback(url)

	o := Options{}
	for _, opt := range opts {
		o.reset()
		opt(&o)
		if len(o.header) != 0 {
			for key, value := range o.header {
				newCallback.Addheader(key, value)
			}
		}
		if len(o.query) != 0 {
			for key, array := range o.query {
				for _, value := range array {
					newCallback.Addparam(key, value)
				}
			}
		}
	}

	var phase *sdk.Phase
	if len(fchain.chain.Phases) == 0 {
		phase = sdk.CreateExecutionPhase()
		fchain.chain.AddPhase(phase)
		emptyPhase = true
	} else {
		phase = fchain.chain.GetLastPhase()
	}
	phase.AddFunction(newCallback)
	return fchain
}

// Apply apply a function with given name and options
// default call is async, provide Sync option to call synchronously
func (fchain *Fchain) Apply(function string, opts ...Option) *Fchain {
	newfunc := sdk.CreateFunction(function)
	sync := false

	o := Options{}
	for _, opt := range opts {
		o.reset()
		opt(&o)
		if len(o.header) != 0 {
			for key, value := range o.header {
				newfunc.Addheader(key, value)
			}
		}
		if len(o.query) != 0 {
			for key, array := range o.query {
				for _, value := range array {
					newfunc.Addparam(key, value)
				}
			}
		}
		if o.sync {
			sync = true
		}
	}

	var phase *sdk.Phase
	if sync {
		if len(fchain.chain.Phases) == 0 {
			phase = sdk.CreateExecutionPhase()
			fchain.chain.AddPhase(phase)
		} else {
			phase = fchain.chain.GetLastPhase()
		}
	} else {
		if emptyPhase {
			phase = fchain.chain.GetLastPhase()
		} else {
			phase = sdk.CreateExecutionPhase()
			fchain.chain.AddPhase(phase)
		}
	}
	emptyPhase = false
	phase.AddFunction(newfunc)

	return fchain
}

// OnFailure set a failure handler routine for the chain
func (fchain *Fchain) OnFailure(handler sdk.ErrorHandler) *Fchain {
	fchain.chain.FailureHandler = handler
	return fchain
}

// Finally set a cleanup handler routine
// it will be called once the execution has finished (Success/Failure)
func (fchain *Fchain) Finally(handler sdk.Handler) *Fchain {
	fchain.chain.Finally = handler
	return fchain
}

// Build encode a underline faaschain (internal)
func (fchain *Fchain) Build() (err error) {
	fchain.chainDef, err = fchain.chain.Encode()
	return err
}

// GetDefinition provide definition of chain (internal)
func (fchain *Fchain) GetDefinition() string {
	return string(fchain.chainDef)
}

// GetChain provides the underlying chain object (internal)
func (fchain *Fchain) GetChain() *sdk.Chain {
	return fchain.chain
}

// GetId returns the current chain id (internal)
func (fchain *Fchain) GetId() string {
	return fchain.id
}

// GetUrl returns the URL for the faaschain function (internal)
func (fchain *Fchain) GetUrl() string {
	return fchain.url
}

// GetAsyncUrl returns the URL for the faaschain async function (internal)
func (fchain *Fchain) GetAsyncUrl() string {
	return fchain.asyncUrl
}
