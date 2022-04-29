# Token Factory

The tokenfactory module allows any account to create a new token with the name `factory/{creator address}/{nonce}
Because tokens are namespaced by creator address, this allows token minting to be permissionless, due to not needing to resolve name collisions.
A single account can create multiple denoms, by providing a unique nonce for each created denom.
Once a denom is created, the original creator is given "admin" privleges over the asset.  This allows them to:
- Mint their denom to any account
- Burn their denom from any account
- Create a transfer of their denom between any two accounts
- Change the admin
In the future, more admin capabilities may be added.  Admins can choose to share admin privledges with other accounts using the authz module. The `ChangeAdmin` functionality, allows changing the master admin account, or even setting it to `""`, meaning no account has admin privledges of the asset.
