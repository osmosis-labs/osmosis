# Osmosis IBC Testing

This package provides helpers, for overriding components of `ibctesting`.

Tracked components we override:
* Adding consensus minimum fees for sent messages
* Adding `SendMsgsNoCheck` as a replacement for `SendMsgs` but without asserting the results as a success. This allows us to test errors.
* Adding a `SignAndDeliver` function as a replacement of `simapp.SignAndDeliver` that does not require an instance of `testing.Testing` and will return the results instead of asserting success or failure. 