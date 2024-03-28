# TSS

## Library selection

| Repo                                                                      | Keygen | Signing | Transport | Security                | Last release | Go version | Notes                  |
|---------------------------------------------------------------------------|--------|---------|-----------|-------------------------|--------------|------------|------------------------|
| [bnb-chain/tss-lib](https://github.com/bnb-chain/tss-lib)                 | ✅      | ✅       | ❌         | Audited on Oct 10, 2019 | Jan 16, 2024 | 1.16       | 705 stars              |
| [thorchain/tss](https://gitlab.com/thorchain/tss/go-tss)                  | ✅      | ✅       | ✅         | Audited on Jun 16, 2020 | Fer 8, 2024  | 1.20       | Production-use example |
| [getamis/alice](https://github.com/getamis/alice)                         | ✅      | ✅       | ❌         | Audited on May 19, 2020 | Nov 30, 2023 | 1.20       | Granted by Coinbase    |
| [taurusgroup/frost-ed25519](https://github.com/taurusgroup/frost-ed25519) | ✅      | ✅       | ❌         | Not audited             | Mar 11, 2021 | 1.14       | Good README            |
| [unit410/threshold-ed25519](https://gitlab.com/unit410/threshold-ed25519) | ✅      | ✅       | ❌         | Not audited             | Feb 21, 2020 | 1.19       |                        |
| [coinbase/kryptology](https://github.com/coinbase/kryptology)             |        |         |           | Papers + HackerOne      | Dec 20, 2021 | 1.17       | Archived               |
| [SwingbyProtocol/tss-lib](https://github.com/SwingbyProtocol/tss-lib)     |        |         |           |                         |              |            | Fork of binance        |

### bnb-chain/tss-lib

Pros:

* Was [audited](https://github.com/bnb-chain/tss-lib?tab=readme-ov-file#security-audit)
on October 10, 2019, by the Kudelski Security
* 700+ stars
* A lot of contributors
* Many libs use it as a basis
* Actively maintained

Cons:

* Doesn't have transport or leader election
* Old Go version

### thorchain/tss

Pros:

* Was [audited](https://kudelskisecurity.com/wp-content/uploads/ThorchainTSSSecurityAudit.pdf) on June 16, 2020, by the
  Kudelski Security
* Has its own transport (!!)
* Actively maintained (11 contributors committing periodically)
* [Production-ready example](https://gitlab.com/thorchain/thornode/-/tree/develop/bifrost/tss)
* Has a built-in leader election

Cons:

* Not popular (11 contributors, 6 stars)
* Doubts on the quality of code (since Throchain itself is quite messy; only doubts without real proofs)

### getamis/alice

Pros:

* Was [audited](https://github.com/getamis/alice?tab=readme-ov-file#audit-report) on May 19, 2020, by the Kudelski
  Security
* 340+ stars
* Actively maintained
* Wide range of cryptographic libs (meaning maintainers know what they are doing)
* Granted by Coinbase

Cons:

* Doesn't have transport or leader election
* HTSS differs from TSS, will need additional time to dig into it