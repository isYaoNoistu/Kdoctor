package runner

type Stage string

const (
	StageLoadConfig           Stage = "load_config"
	StageResolveProfile       Stage = "resolve_profile"
	StageParseInputs          Stage = "parse_inputs"
	StageCollectBaseSnapshots Stage = "collect_base_snapshots"
	StageCollectKafka         Stage = "collect_kafka_snapshots"
	StageRunChecks            Stage = "run_checks"
	StageDiagnose             Stage = "diagnose"
	StageRender               Stage = "render"
	StageExit                 Stage = "exit"
)
