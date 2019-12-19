module github.com/ldsec/drynx

go 1.12

// fix ID discrepancies between typescript and golang
replace go.dedis.ch/kyber/v3 => go.dedis.ch/kyber/v3 v3.0.12-0.20191209094922-c336cade8388

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/alex-ant/gomath v0.0.0-20160516115720-89013a210a82
	github.com/cdipaolo/goml v0.0.0-20190412180403-e1f51f713598
	github.com/coreos/bbolt v1.3.3
	github.com/fanliao/go-concurrentMap v0.0.0-20141114143905-7d2d7a5ea67b
	github.com/ldsec/unlynx v1.4.0
	github.com/montanaflynn/stats v0.5.0
	github.com/pelletier/go-toml v1.6.0
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/tonestuff/quadratic v0.0.0-20141117024252-b79de8af2377
	github.com/urfave/cli v1.22.2
	go.dedis.ch/cothority/v3 v3.4.0
	go.dedis.ch/kyber/v3 v3.0.11
	go.dedis.ch/onet/v3 v3.0.31
	go.dedis.ch/protobuf v1.0.11
	golang.org/x/crypto v0.0.0-20191219195013-becbf705a915
	gonum.org/v1/gonum v0.6.1
	gopkg.in/satori/go.uuid.v1 v1.2.0
)
