# Txfees

This module allows validators to define their min gas price is a single "base denom", but then allows users to define their tx fees in any whitelisted fee token.  It does this by converting the whitelisted fee token to its equivalent value in base denom fee, using a "Spot Price calculator" (such as the gamm keeper).