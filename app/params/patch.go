package params

// V32PatchActivationHeight is the first osmosis-1 height at which the v32.0.1
// consensus changes apply. Below it, v32.0.1 executes blocks exactly as
// v32.0.0, so a v32.0.1 node replays history unchanged and validators can
// switch binaries at any time before this height.
//
// 1,000,000,000 is a placeholder that keeps every change inactive. It must be
// set to the height agreed with validators before the release is tagged.
const V32PatchActivationHeight int64 = 1_000_000_000
